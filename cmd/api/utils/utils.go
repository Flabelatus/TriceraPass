// Package utils provides utility functions for common operations
// such as reading and writing JSON responses, error handling, and other
// helper methods that are reused across different parts of the application.
package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteJSON marshals a payload to JSON and writes it to the HTTP response writer.
// It sets the provided status code and allows for optional HTTP headers to be added.
// If there is an error during marshalling or writing, it returns the error.
//
// Parameters:
// - w: The HTTP response writer to write the JSON response to.
// - status: The HTTP status code to set in the response.
// - payload: The data to be marshaled into JSON.
// - headers (optional): HTTP headers to include in the response.
//
// Returns:
// - error: An error if marshalling or writing to the response fails.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
	out, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// ReadJSON reads a JSON request body and decodes it into the provided payload.
// It ensures the request body does not exceed a size limit and rejects unknown fields.
//
// Parameters:
// - w: The HTTP response writer (used to handle errors, if any).
// - r: The HTTP request containing the JSON body to be decoded.
// - payload: The structure to decode the JSON data into.
//
// Returns:
// - error: An error if decoding fails, if the body exceeds the size limit, or if there are unknown fields.
func ReadJSON(w http.ResponseWriter, r *http.Request, payload interface{}) error {
	maxBytes := 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(payload)
	if err != nil {
		return err
	}
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only a single JSON value")
	}
	return nil
}

// ErrorJSON sends a JSON response with an error message.
// It defaults to HTTP status 400 (Bad Request) unless a different status code is provided.
// The error message is included in the response.
//
// Parameters:
// - w: The HTTP response writer to write the JSON error response to.
// - err: The error to include in the response.
// - status (optional): An optional HTTP status code to set in the response (defaults to 400).
//
// Returns:
// - error: An error if writing the JSON response fails.
func ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	if err != nil {
		payload.Message = err.Error()
	}

	return WriteJSON(w, statusCode, payload)
}
