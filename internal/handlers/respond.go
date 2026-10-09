package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeStatusJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, payload any) {
	writeStatusJSON(w, http.StatusOK, payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeStatusJSON(w, status, errorResponse{Error: message})
}
