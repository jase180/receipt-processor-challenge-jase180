package store

import (
	"errors"
	"sync"

	"receipt-processor-challenge-jase180/internal/models"
)

// Defined errors for reusability
var (
	ErrReceiptAlreadyExists = errors.New("receipt already exists in database")
	ErrReceiptNotInDatabase = errors.New("no such receipt exists in database")
)

// MemoryDatabase provides an in-memory storage for receipts
// Use sync.RWMutex to ensure write safety (sync.Map is alternative)
type MemoryDatabase struct {
	lock     sync.RWMutex              // lock ensures thread safety
	receipts map[string]models.Receipt // Stores receipts in memory
}

// NewMemoryDatabase initializes and returns a new in-memory database
func NewMemoryDatabase() *MemoryDatabase {
	db := &MemoryDatabase{}                       // initiates a db
	db.receipts = make(map[string]models.Receipt) // makes a map with the Receipt() struct from models

	return db
}

// AddReceipt adds a receipt into the memory database after checking if a receipt with the same ID exists already
func (db *MemoryDatabase) AddReceipt(receipt models.Receipt) error {
	// Manual lock/unlock to ensure go concurrency, only one goroutine allowed access at a time
	db.lock.Lock()
	defer db.lock.Unlock()

	// Check if receipt for ID exists already
	_, exists := db.receipts[receipt.ID]
	if exists {
		return ErrReceiptAlreadyExists
	}

	// Add receipt into the MemoryDatabase
	db.receipts[receipt.ID] = receipt
	return nil
}

// GetReceiptByID retrieves the receipt from the memory database with the ID after checking if ID exists
func (db *MemoryDatabase) GetReceiptByID(id string) (models.Receipt, error) {
	// Manual lock/unlock to ensure go concurrency, only one goroutine allowed access at a time
	db.lock.RLock()
	defer db.lock.RUnlock()

	// Retrieve receipt with ID
	receipt, exists := db.receipts[id]
	if !exists {
		return models.Receipt{}, ErrReceiptNotInDatabase
	}

	return receipt, nil
}
