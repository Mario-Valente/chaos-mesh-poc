package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var backendURL string

func init() {
	backendURL = os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://backend-service:8081"
	}
}

func main() {
	log.Println("Frontend Service starting on :8080")
	log.Printf("Backend URL: %s\n", backendURL)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/data", dataHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"frontend"}`)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Frontend] Received request: %s %s\n", r.Method, r.RequestURI)

	// Create a client with timeout to demonstrate chaos effects
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Call backend service
	backendEndpoint := backendURL + "/api/process"
	log.Printf("[Frontend] Calling backend: %s\n", backendEndpoint)

	resp, err := client.Get(backendEndpoint)
	if err != nil {
		log.Printf("[Frontend] Error calling backend: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"backend unavailable","details":"%s"}`, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Frontend] Error reading backend response: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"failed to read backend response"}`)
		return
	}

	// Return backend response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
	log.Printf("[Frontend] Response sent successfully\n")
}
