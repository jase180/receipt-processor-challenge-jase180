package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"receipt-processor-challenge-jase180/internal/models"
	"receipt-processor-challenge-jase180/internal/store"
)

func TestCreateReceiptHandler(t *testing.T) {
	// Initialize database and handler
	db := store.NewMemoryDatabase()
	handler := NewReceiptHandler(db)

	tests := []struct {
		name         string
		receipt      models.Receipt
		responseCode int  // corresponding response codes
		wantID       bool // true if expect success
	}{
		{
			name: "Valid Receipt from README",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusOK,
			wantID:       true,
		},
		{
			name: "Duplicate Receipt",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusOK,
			wantID:       true,
		},
		{
			name:         "Empty struct",
			receipt:      models.Receipt{},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		},
		{
			name: "Missing Retailer",
			receipt: models.Receipt{
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		},
		{
			name: "Missing date",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		},
		{
			name: "Missing time",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		},
		{
			name: "Missing total",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Missing entire items",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Total:        "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Missing price in items",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Missing shortDescription in items",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad date format",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "20220101",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad time format",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "1301",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad total format (negative)",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "-35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad total format (not 2 dp)",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.3535",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad price format (negative)",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "-12.25"}, // negative here
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		}, {
			name: "Bad price format (not 2 dp)",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.4956"}, // not 2 dp here
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			responseCode: http.StatusBadRequest,
			wantID:       false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var body []byte

			// Marshal the test receipt
			body, err := json.Marshal(testCase.receipt)
			if err != nil {
				t.Fatalf("Error marshaling test receipt to JSON: %v", err)
			}

			result := httptest.NewRequest("POST", "/receipts/process", bytes.NewReader(body))
			result.Header.Set("Content-Type", "application/json")

			// Create and response recorder and call handler
			responseRecorder := httptest.NewRecorder()
			handler.CreateReceiptHandler(responseRecorder, result)

			// Check if it has correct response code
			if responseRecorder.Code != testCase.responseCode {
				t.Errorf("Result status: %d, want: %d", responseRecorder.Code, testCase.responseCode)
			}

			// Check if it has correct response ID by unmarshaling
			var response map[string]string
			err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Error during test parsing successful result JSON: %v", err)
			}

			if testCase.wantID {
				_, exists := response["id"]
				if !exists {
					t.Errorf("No ID response despite successful 200 response")
				}
			} else {
				_, exists := response["id"]
				if exists {
					t.Errorf("ID response despite not expecting one")
				}
			}
		})
	}

}

func TestGetReceiptHandler(t *testing.T) {
	// Initialize database and handler
	db := store.NewMemoryDatabase()
	handler := NewReceiptHandler(db)

	// Create a receipt and add to the database to test getting (using README example)
	testID := uuid.NewString() // Generate ID for test receipt
	testReceiptTargetREADME := models.Receipt{
		ID:           testID,
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []models.Item{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
			{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
			{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
			{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
			{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
		},
		Total: "35.35",
	}

	// Directly add test receipt to database to avoid POST error
	db.AddReceipt(testReceiptTargetREADME)

	tests := []struct {
		name         string
		receiptID    string
		responseCode int  // corresponding response codes
		wantPoints   bool // true if expect success

	}{
		{"Valid ID and receipt", testID, http.StatusOK, true},
		{"Valid ID and no such receipt", uuid.NewString(), http.StatusNotFound, false},
		{"Invalid ID", "ABCDEFG", http.StatusBadRequest, false},
		{"Empty ID", "", http.StatusBadRequest, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create request
			result := httptest.NewRequest("GET", "/receipts/"+testCase.receiptID+"/points", nil)

			// Create and response recorder and call handler
			responseRecorder := httptest.NewRecorder()

			// Inject the route
			result = mux.SetURLVars(result, map[string]string{
				"id": testCase.receiptID,
			})
			//
			handler.GetReceiptHandler(responseRecorder, result)

			// Check if it has correct response code
			if responseRecorder.Code != testCase.responseCode {
				t.Errorf("Result status: %d, want: %d", responseRecorder.Code, testCase.responseCode)
			}

			// Check if it has correct response ID by unmarshaling
			var response map[string]interface{}
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
			if err != nil && responseRecorder.Code == http.StatusOK {
				t.Fatalf("Error during test parsing successful result JSON: %v", err)
			}

			// Check if it has correct response ID by unmarshaling
			if testCase.wantPoints {
				_, exists := response["points"]
				if !exists {
					t.Errorf("No points in response despite expecting them")
				}
			} else {
				// For error cases, we might want to check for error message
				if responseRecorder.Code != http.StatusOK {
					_, exists := response["error"]
					if !exists && len(responseRecorder.Body.Bytes()) > 0 {
						t.Errorf("Expected error message for failed response")
					}
				}
			}
		})
	}

}
