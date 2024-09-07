package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (app *application) writeJSON(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, payload interface{}) error {
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

func (app *application) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	if err != nil {
		payload.Message = err.Error()
	}

	return app.writeJSON(w, statusCode, payload)
}

// Load settings.yml and parse into the Config struct
func loadConfig(configFile string) (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return config, nil
}

// Helper function to check if the origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
