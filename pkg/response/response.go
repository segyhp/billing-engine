package response

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response := Response{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      data,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Success sends a successful JSON response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created sends a created JSON response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// Error sends an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string, err error) {
	response := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("Error encoding error response: %v", encodeErr)
	}
}

// BadRequest sends a 400 bad request response
func BadRequest(w http.ResponseWriter, message string, err error) {
	Error(w, http.StatusBadRequest, message, err)
}

// NotFound sends a 404 not found response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, nil)
}

// InternalServerError sends a 500 internal server error response
func InternalServerError(w http.ResponseWriter, message string, err error) {
	Error(w, http.StatusInternalServerError, message, err)
}

// Unauthorized sends a 401 unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, nil)
}

// JSONMiddleware sets JSON content type for all responses
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response recorder to capture the status code
		recorder := &responseRecorder{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, recorder.statusCode, duration)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
