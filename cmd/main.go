package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"receipt-processor-challenge-jase180/internal/handlers"
	"receipt-processor-challenge-jase180/internal/store"
)

// main initializes the in-memory database, sets up routes and starts the server
func main() {
	// Initialize in-memory database and handler
	db := store.NewMemoryDatabase()
	handler := handlers.NewReceiptHandler(db)

	// Create Router with gorilla/mux over just using net/http to grab dynamic link ID for GET easily
	router := mux.NewRouter()

	// POST /receipts/process
	// Accepts Receipt JSON object and stores in memory database
	// Returns 200 and generated UUID for created receipt if successful
	// Returns 400 and bad request if unsuccessful
	router.HandleFunc("/receipts/process", handler.CreateReceiptHandler).Methods(http.MethodPost)

	// GET /receipts/{id}/points
	// Returns 200 and points for requested receipt if successful
	// Returns 400 and bad request if unsuccessful
	router.HandleFunc("/receipts/{id}/points", handler.GetReceiptHandler).Methods(http.MethodGet)

	// Start the server
	port := ":8080" // start the server on port 8080 for local development
	log.Println("Running local server: " + port)
	log.Fatal(http.ListenAndServe(port, router))
}
