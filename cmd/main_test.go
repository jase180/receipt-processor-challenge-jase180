package main

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"

	"receipt-processor-challenge-jase180/internal/handlers"
	"receipt-processor-challenge-jase180/internal/store"
)

// Integration test with two receipts
// Receipts added together to test database can handle more than one test
func TestIntegration(t *testing.T) {

	// Reimplement server and router in main.go for testing for control and isolation
	db := store.NewMemoryDatabase()
	handler := handlers.NewReceiptHandler(db)

	router := mux.NewRouter()

	router.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request) {
		handler.CreateReceiptHandler(w, r)
	}).Methods(http.MethodPost)

	router.HandleFunc("/receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		handler.GetReceiptHandler(w, r)
	}).Methods(http.MethodGet)

	// Start test server with httptest
	server := httptest.NewServer(router)
	defer server.Close() // proper clean up

	//  Receipts JSON of all all given examples
	targetReceipt := `{
		"retailer": "Target",
		"purchaseDate": "2022-01-01",
		"purchaseTime": "13:01",
		"items": [
		  {
			"shortDescription": "Mountain Dew 12PK",
			"price": "6.49"
		  },{
			"shortDescription": "Emils Cheese Pizza",
			"price": "12.25"
		  },{
			"shortDescription": "Knorr Creamy Chicken",
			"price": "1.26"
		  },{
			"shortDescription": "Doritos Nacho Cheese",
			"price": "3.35"
		  },{
			"shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
			"price": "12.00"
		  }
		],
		"total": "35.35"
	}`

	mmCornerReceipt := `{
		"retailer": "M&M Corner Market",
		"purchaseDate": "2022-03-20",
		"purchaseTime": "14:33",
		"items": [
		  {
			"shortDescription": "Gatorade",
			"price": "2.25"
		  },{
			"shortDescription": "Gatorade",
			"price": "2.25"
		  },{
			"shortDescription": "Gatorade",
			"price": "2.25"
		  },{
			"shortDescription": "Gatorade",
			"price": "2.25"
		  }
		],
		"total": "9.00"
	}`

	receiptsArray := []string{targetReceipt, mmCornerReceipt}
	responseIDArray := []string{}

	// Send POST request for all receipts
	for _, receipt := range receiptsArray {
		// Send POST
		response, err := http.Post(server.URL+"/receipts/process", "application/json", bytes.NewBuffer([]byte(receipt)))
		if err != nil {
			t.Errorf("Failed to send POST: %v for %v", err, receipt)
		}
		defer response.Body.Close()

		// Check status code
		if response.StatusCode != http.StatusOK {
			t.Errorf("Result: %d, want 200", response.StatusCode)
		}

		// Extract ID to responseIDArray from response
		var responseJSON map[string]string
		body, _ := io.ReadAll(response.Body)
		json.Unmarshal(body, &responseJSON)
		receiptID, exists := responseJSON["id"]
		if !exists {
			t.Errorf("Response did not contain 'id' for %v", receipt)
		}
		responseIDArray = append(responseIDArray, receiptID)
	}

	responsePointsArray := []int{}

	// Fetch points from all receipts
	for _, id := range responseIDArray {
		response, err := http.Get(server.URL + "/receipts/" + id + "/points")
		if err != nil {
			t.Errorf("Failed to send GET: %v for %v", err, id)
		}
		defer response.Body.Close()

		// Check status code
		if response.StatusCode != http.StatusOK {
			t.Errorf("Result: %d, want 200", response.StatusCode)
		}

		// Extract ID to responseIDArray from response
		var responseJSON map[string]int
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()

		json.Unmarshal(body, &responseJSON)
		receiptPoints, exists := responseJSON["points"]
		if !exists {
			t.Errorf("Response did not contain points for %v", id)
		}

		responsePointsArray = append(responsePointsArray, receiptPoints)
	}

	// Verify points
	wantPointsArray := []int{28, 109}
	for i, points := range wantPointsArray {
		if points != responsePointsArray[i] {
			t.Errorf("Wanted %v points, got %v", wantPointsArray[i], responsePointsArray[i])
		}
	}
}

// Smoke test - starting server simulating real execution using a goroutine
// Only testing a post since purpose is to make sure main will run locally
func TestSmoke(t *testing.T) {
	go func() {
		main()
	}()
	time.Sleep(3 * time.Second) // give 3 seconds for it to start, should be more than enough

	simpleReceipt := `{
		"retailer": "Target",
		"purchaseDate": "2022-01-02",
		"purchaseTime": "13:13",
		"total": "1.25",
		"items": [
			{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}
		]
	}`

	response, err := http.Post("http://localhost:8080/receipts/process", "application/json", bytes.NewBuffer([]byte(simpleReceipt)))
	if err != nil {
		t.Errorf("Failed to send POST: %v for %v", err, simpleReceipt)
	}
	defer response.Body.Close()

	// Check status code
	if response.StatusCode != http.StatusOK {
		t.Errorf("Result: %d, want 200", response.StatusCode)
	}

}
