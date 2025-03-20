package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"receipt-processor-challenge-jase180/internal/handlers"
	"receipt-processor-challenge-jase180/internal/store"
)

// main also sets up router since this is a simple webservice
func main() {
	// Initialize in-memory database and handler
	db := store.NewMemoryDatabase()
	handler := handlers.NewReceiptHandler(db)

	// Create Router with gorilla/mux
	// gorilla/mux used over just using net/http to grab dynamic link ID for GET
	router := mux.NewRouter()

	// POST, Takes in Receipt JSON object and adds to in memory database
	//returns 200 and generated UUID for created receipt if successful
	// returns 400 and bad request if unsuccessful
	router.HandleFunc("/receipts/process", handler.CreateReceiptHandler).Methods(http.MethodPost)

	// GET,
	// Returns 200 and points for requested receipt if successful
	// Returns 400 and bad request if unsuccessful
	router.HandleFunc("/receipts/{id}/points", handler.GetReceiptHandler).Methods(http.MethodGet)

	// Start the server
	port := ":8080" // port 8080 is go convention
	log.Println("Running local server: " + port)
	log.Fatal(http.ListenAndServe(port, router))
}
