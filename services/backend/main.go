package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var dbConnString string

func init() {
	dbConnString = os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		dbConnString = "postgres://postgres:postgres@postgres-service:5432/chaos_db?sslmode=disable"
	}
}

func main() {
	log.Println("Backend Service starting on :8081")
	log.Printf("Database URL: %s\n", dbConnString)

	// Connect to database with retries
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dbConnString)
		if err != nil {
			log.Printf("Error opening database (attempt %d): %v\n", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Error pinging database (attempt %d): %v\n", i+1, err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Connected to database successfully")
		break
	}

	if err != nil {
		log.Printf("Failed to connect to database after retries: %v\n", err)
		// Continue anyway for demo purposes
	} else {
		defer db.Close()
	}

	// Initialize database schema
	initDB()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/process", processHandler)
	http.HandleFunc("/api/stats", statsHandler)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func initDB() {
	if db == nil {
		log.Println("Database not connected, skipping schema initialization")
		return
	}

	schema := `
	CREATE TABLE IF NOT EXISTS requests (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status VARCHAR(50),
		response_time_ms INTEGER
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error initializing schema: %v\n", err)
	} else {
		log.Println("Database schema initialized")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"backend"}`)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[Backend] Processing request\n")

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Try to write to database
	status := "success"
	if db != nil {
		query := `
		INSERT INTO requests (status, response_time_ms)
		VALUES ($1, $2)
		`
		_, err := db.Exec(query, status, time.Since(startTime).Milliseconds())
		if err != nil {
			log.Printf("[Backend] Error writing to database: %v\n", err)
			status = "db_error"
		}
	} else {
		log.Println("[Backend] Database not connected, skipping write")
		status = "no_db"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"result":"processed","status":"%s","duration_ms":%d}`, status, time.Since(startTime).Milliseconds())
	log.Printf("[Backend] Request completed in %dms\n", time.Since(startTime).Milliseconds())
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"database not connected"}`)
		return
	}

	var count int
	var avgTime int64
	query := `SELECT COUNT(*), COALESCE(AVG(response_time_ms), 0) FROM requests;`
	err := db.QueryRow(query).Scan(&count, &avgTime)
	if err != nil {
		log.Printf("[Backend] Error querying stats: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"failed to query stats"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"total_requests":%d,"avg_response_time_ms":%d}`, count, avgTime)
}
