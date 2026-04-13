package logger

import (
	"log"
)

// ❌ PROBLEM: Unstructured logging
// ❌ Hard to parse in production
// ❌ Difficult to correlate across services
// ❌ No correlation IDs
// ❌ No structured fields

func LogError(message string, err error) {
	// ❌ Unstructured: Hard to parse
	log.Printf("ERROR: %s - %v", message, err)
	// Output: "ERROR: Payment processing failed - connection timeout"
	// ❌ Can't easily extract fields for analysis
	// ❌ Difficult to correlate with distributed tracing
}

func LogInfo(message string) {
	// ❌ Unstructured: No context
	log.Println(message)
	// Output: "Order created"
	// ❌ Missing: which order? when? by which user?
}

// ❌ What we're missing:
// - Correlation IDs (request tracing)
// - Structured JSON output
// - Log levels (DEBUG, INFO, WARN, ERROR)
// - Contextual fields (userId, orderId, requestId)
// - Timestamps
// - Service name
// - Environment

// ✅ What it should look like (commented out):
/*
import "github.com/sirupsen/logrus"

var log = logrus.New()

func LogErrorStructured(correlationID string, userID string, message string, err error) {
	log.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"user_id": userID,
		"error": err.Error(),
		"timestamp": time.Now(),
	}).Error(message)

	// Output (JSON):
	// {"correlation_id":"req-123","user_id":"user-456","error":"connection timeout","timestamp":"2024-04-13T10:30:00Z","level":"error","msg":"Payment processing failed"}
}
*/
