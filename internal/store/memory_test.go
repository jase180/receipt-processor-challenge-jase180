package store

import (
	"sync"
	"testing"

	"github.com/google/uuid"

	"receipt-processor-challenge-jase180/internal/models"
)

// TestMemoryDataBaseCreation tests basic memory database creation
func TestMemoryDataBaseCreation(t *testing.T) {
	db := NewMemoryDatabase()

	if db.receipts == nil {
		t.Fatalf("Receipt map not found, want MemoryDatabase initiated correctly")
	}
}

// TestMemoryDatabaseFunctions tests all basic functions and errors for the Memory Database
func TestMemoryDatabaseFunctions(t *testing.T) {
	db := NewMemoryDatabase()

	// given example: morning-receipt
	receiptMorning := models.Receipt{
		ID:           uuid.NewString(), // Generate a new id with google/uuid
		Retailer:     "Walgreens",
		PurchaseDate: "2022-01-02",
		PurchaseTime: "08:13",
		Total:        "2.65",
		Items: []models.Item{
			{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
			{ShortDescription: "Dasani", Price: "1.40"},
		},
	}

	// given example: simple-receipt
	receiptSimple := models.Receipt{
		ID:           uuid.NewString(),
		Retailer:     "Target",
		PurchaseDate: "2022-01-02",
		PurchaseTime: "13:13",
		Total:        "1.25",
		Items: []models.Item{
			{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
		},
	}

	// Test AddReceipt to empty database
	err := db.AddReceipt(receiptMorning)
	if err != nil {
		t.Fatalf("Result: %v; want Success Add", err)
	}

	// Test AddReceipt to non-empty database
	err = db.AddReceipt(receiptSimple)
	if err != nil {
		t.Fatalf("Result: %v; want Success Add", err)
	}

	// Test AddReceipt for ID exists already - ErrReceiptAlreadyExists
	err = db.AddReceipt(receiptMorning)
	if err != ErrReceiptAlreadyExists {
		t.Fatalf("Result: %v; want error %v", err, ErrReceiptAlreadyExists)
	}

	// Test GetReceiptByID for first receipt
	_, err = db.GetReceiptByID(receiptMorning.ID)
	if err != nil {
		t.Fatalf("Result: %v; want Success Retrieve", err)
	}

	// Test GetReceiptByID for No such ID - ErrReceiptNotInDatabase
	_, err = db.GetReceiptByID(uuid.NewString()) // Generate a new id with google/uuid
	if err != ErrReceiptNotInDatabase {
		t.Fatalf("Result: %v; want error %v", err, ErrReceiptNotInDatabase)
	}

	// Test correct receipt retrieved
	testReceiptMorning, err := db.GetReceiptByID(receiptMorning.ID)
	if err != nil {
		t.Fatalf("Result: %v; want Success Retrieve", err)
	}
	if testReceiptMorning.Retailer != receiptMorning.Retailer {
		t.Fatalf("Incorrect receipt (retailer) retrieved, Result: %v; want %v", testReceiptMorning.Retailer, receiptMorning.Retailer)
	}
}

// Tests concurrency with WaitGroup to read and write at the same time
func TestMemoryDatabaseConcurrency(t *testing.T) {
	db := NewMemoryDatabase()

	// iven example: morning-receipt
	receiptMorning := models.Receipt{
		ID:           uuid.NewString(), // Generate a new id with google/uuid
		Retailer:     "Walgreens",
		PurchaseDate: "2022-01-02",
		PurchaseTime: "08:13",
		Total:        "2.65",
		Items: []models.Item{
			{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
			{ShortDescription: "Dasani", Price: "1.40"},
		},
	}

	// Add one receipt for reading
	if err := db.AddReceipt(receiptMorning); err != nil {
		t.Fatalf("Result: %v; want Success Add", err)
	}

	// Initiate a waitgroup
	var waitGroup sync.WaitGroup
	numConcurrentTasks := 10 // number of goroutines that will try to clash (10 is arbitrary)

	// Concurrent reads
	for i := 0; i < numConcurrentTasks; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done() // registers the go routine even if there is an error

			// Read by GetReceiptByID
			_, err := db.GetReceiptByID(receiptMorning.ID)
			if err != nil {
				t.Errorf("Error in concurrent READ: %v", err)
			}
		}()
	}

	// Concurrent writes of new example receipts (only difference is ID)
	for i := 0; i < numConcurrentTasks; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			concurrentWriteReceipt := models.Receipt{
				ID:           uuid.NewString(),
				Retailer:     "Target",
				PurchaseDate: "2022-01-02",
				PurchaseTime: "13:13",
				Total:        "1.25",
				Items: []models.Item{
					{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
				},
			}
			err := db.AddReceipt(concurrentWriteReceipt)
			if err != nil {
				t.Errorf("Error in concurrent WRITE: %v", err)
			}
		}()
	}

	waitGroup.Wait() // this ensures all go routines finish
}
