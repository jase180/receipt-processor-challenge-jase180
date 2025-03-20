package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"receipt-processor-challenge-jase180/internal/models"
	rules "receipt-processor-challenge-jase180/internal/services"
	"receipt-processor-challenge-jase180/internal/store"
)

// A struct that creates connection to database
type ReceiptHandler struct {
	Database *store.MemoryDatabase
}

// NewReceiptHandler creates a new handler that connects to existing database
// Panic because database is critical.  Error less preferred because webservice requires database
func NewReceiptHandler(db *store.MemoryDatabase) *ReceiptHandler {
	if db == nil {
		panic("Database does not exist.  Cannot initialize.")
	}
	return &ReceiptHandler{Database: db}
}

// helper function that takes errors and encode it into a JSON
func sendJSON(w http.ResponseWriter, message interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	// Marshal the message first to catch errors and avoid sending faulty JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		http.Error(w, `{"message": "JSON marshaling error"}`, http.StatusInternalServerError)
		return
	}

	// Write the JSON response
	w.Write(jsonMessage)
}

// GetReceiptHandler takes a GET request with /receipts/{id}/points endpoint, where dynamic id is a UUID for a receipt
// Validates JSON format, ID format, and if ID is in database
func (h *ReceiptHandler) GetReceiptHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve ID required using gorilla/mux or alternatively query with id := r.URL.Query().Get("id")
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if id is not empty in JSON and raise error again, even though router in main also checks
	if id == "" {
		sendJSON(w, map[string]string{"error": "BadRequest: No ID given in query"}, http.StatusBadRequest) //  400 response
		return
	}

	// Check valid UUID format (generated from google/uuid)
	if _, err := uuid.Parse(id); err != nil {
		sendJSON(w, map[string]string{"error": "BadRequest: Invalid ID format"}, http.StatusBadRequest)

		return
	}

	// Look up ID and raise error if no ID found
	receipt, err := h.Database.GetReceiptByID(id)
	if err != nil {
		sendJSON(w, map[string]string{"error": "No receipt found for that ID"}, http.StatusNotFound) // 404 response
		return
	}

	// Calculate points by calling rules.go
	points := rules.CalculatePoints(receipt)

	// Create calculated points response
	response := map[string]int{
		"points": points,
	}

	// Set status to 200 OK meaning success and send
	sendJSON(w, response, http.StatusOK)
}

// CreateReceiptHandler validates incoming POST JSON object and writes to in memory database
// Validations include JSON, receipt structure, DDoS and resource exhaustion prevention
// Assumptions: Identical duplicate receipts allowed
func (h *ReceiptHandler) CreateReceiptHandler(w http.ResponseWriter, r *http.Request) {
	// Size limiting to prevent DoS and resource exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	defer r.Body.Close()                           // Proper clean up

	// Read the request body before decoding to debug
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendJSON(w, map[string]string{"error": "Invalid request body"}, http.StatusBadRequest)
		return
	}

	// Create empty receipt struct
	var receipt models.Receipt

	// Unmarshal JSON into the Receipt struct, only ID missing now, error if invalid JSON
	err = json.Unmarshal(bodyBytes, &receipt)
	if err != nil {
		sendJSON(w, map[string]string{"error": "Invalid JSON"}, http.StatusBadRequest) // 400 response
		return
	}

	// Validate JSON contains required fields using helper function
	if err := validateReceipt(receipt); err != nil {
		sendJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest) // 400
		return
	}

	// Generate new UUID for receipt, now receipt model struct completely filled
	newID := uuid.New().String()
	receipt.ID = newID

	// Add receipt to memory database, and error if failure
	createErr := h.Database.AddReceipt(receipt)
	if createErr != nil {
		sendJSON(w, map[string]string{"error": "Database failure, could not create receipt"}, http.StatusInternalServerError) // 500 response
		return
	}

	// Create new receipt ID response
	response := map[string]string{
		"id": newID,
	}

	// Set status to 200 OK meaning success and send
	sendJSON(w, response, http.StatusOK)
}

// Helper function verifying Receipt structure and data type fits openAPI
// Empty string checks, and then format checks
func validateReceipt(receipt models.Receipt) error {

	//check if retailer is non empty string
	if strings.TrimSpace(receipt.Retailer) == "" {
		return errors.New("BadRequest: The receipt is invalid. Retailer string is empty")
	}
	//check if date is non empty string
	if strings.TrimSpace(receipt.PurchaseDate) == "" {
		return errors.New("BadRequest: The receipt is invalid. Purchase date string is empty")
	}
	//check if time is non empty string
	if strings.TrimSpace(receipt.PurchaseTime) == "" {
		return errors.New("BadRequest: The receipt is invalid. Purchase time string is empty")
	}
	//check if total is non empty string
	if strings.TrimSpace(receipt.Total) == "" {
		return errors.New("BadRequest: The receipt is invalid. Total string is empty")
	}
	//check if item has at least 1 item
	if len(receipt.Items) == 0 {
		return errors.New("BadRequest: The receipt is invalid. no items found")
	}
	//check for each item in Items has shortDescription and price non empty string
	for _, item := range receipt.Items {
		if strings.TrimSpace(item.ShortDescription) == "" {
			return errors.New("BadRequest: The receipt is invalid. Item short description string is empty")
		}
		if strings.TrimSpace(item.Price) == "" {
			return errors.New("BadRequest: The receipt is invalid. Item Price string is empty")
		}
	}

	// Date, Time, Total, Price format checks
	// Check date format
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return errors.New("BadRequest: The receipt is invalid. Receipt date format is incorrect")
	}

	// Check time format
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return errors.New("BadRequest: The receipt is invalid. Receipt time format is incorrect")
	}

	// Check Total format - 2 digits, non negative (assume 0 dollars allowed)
	var regexTotalDollar = regexp.MustCompile(`^\d+\.\d{2}$`)
	if !regexTotalDollar.MatchString(receipt.Total) {
		return errors.New("BadRequest: The receipt is invalid. Receipt Total format is incorrect")
	}

	// Check item.Price format - 2 digits, non negative (assume 0 dollars allowed)
	var regexItemPriceDollar = regexp.MustCompile(`^\d+\.\d{2}$`)
	for _, item := range receipt.Items {
		if !regexItemPriceDollar.MatchString(item.Price) {
			return errors.New("BadRequest: The receipt is invalid. Item price format is incorrect")
		}
	}

	// If no errors
	return nil
}
