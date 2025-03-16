# Golang Web Service Design Document

## Project Name: Fetch rewards receipt processor challenge
## Author: Jason Lee

---

## 1. Overview
This project involves building a receipt processor webservice that fulfills API documented in api.yml.  The API includes two Endpoints, one GET and one POST in JSON. Two types: receipt and item.  POST will take receipt JSON which is to be stored in in-memory database and should return the generated UUID. GET will process dynamic ID, access corresponding receipt, calculate points from given rules and return points.  This is to be packaged in docker.

---

## 2. Goals & Non-Goals

### Goals
- Implement a simple, lightweight RESTful API.
- Use the standard `net/http` package for better understanding of Golang standard packages. ('gin' would be an alternative)
    - Use Gorilla/mux to process dynamic query
    - Use google/uuid to generate uuid
- Input validation
- Create functional in memory database
- Calculate correct points
- Implement unit tests and integration tests.
- Implement via Docker
- Provide instructions to installation and usage

### Non-Goals
- No frontend
- No database, only in memory database

---

## 3. Architecture

### High-Level Design
```plaintext
/receipt-app
│
├── cmd/
│   └── main.go              # Main entry point of the app (starting the server)
│
├── handlers/
│   └── handlers.go          # Handlers for POST and GET API
│
├── models/
│   └── models.go            # Struct for Receipt and Item
│
├── services/
│   └── rules.go             # Business logic to calculate points for GET
│
├── store/
│   └── memory.go            # In memory storage
│
└── go.mod                   # Go module file
└── go.sum                   # Go dependencies checksum


```

---

## 4. API Endpoints

| POST  | `/receipts/process`        | Accepts JSON input, stores in in memory and returns a generated UUID. 
| GET   | `/receipts/{id}/points`    | Fetches the receipt by {id}, calculates points, and returns the computed points. 

---

## 5. Component Breakdown

### main (`main.go`)
- Uses `net/http` with `gorilla/mux` to define API routes.
- Maps endpoints to corresponding handlers.
- Routing in main.go instead of routers.go for simplicity because only 2 methods
    - Can move/expand to a routers.go if scope change

### Handlers (`handlers.go`)
- Implements logic for processing incoming requests.
- Returns structured JSON responses.
- middleware like validation implemented in handlers.go as well for simplicity because of small project scope
    - Can move/expand to have middleware.go as well if scope change

### models (`models.go`)
- Contains structs for receipt and memory
- Uses structs rather than interface because only in memory storage required

### Memory (`memory.go`)
- Contains struct and methods for initializing an in memory database
- Contains methods for adding and retrieving receipts
    - Does not contain methods for editing or removing because not in scope

### Rules (`rules.go`)
- Functions for calculating points

---

## 6. Testing Strategy

### Unit Testing
### Integration Testing
### Concurrency Testing

---

## 7. Security Considerations

- Validate all input - prevent injection attacks.  

---

## 8. Deployment

### Deployment Options
- Local development using `go run main.go`  
- Docker containerized deployment  

---