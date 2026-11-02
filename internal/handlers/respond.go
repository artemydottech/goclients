package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeStatusJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response failed", "err", err)
	}
}

func writeJSON(w http.ResponseWriter, payload any) {
	writeStatusJSON(w, http.StatusOK, payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeStatusJSON(w, status, errorResponse{Error: message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	body, err := io.ReadAll(r.Body)
	if err == nil {
		if bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			err = errors.New("request body cannot be null")
		} else {
			err = json.Unmarshal(body, dst)
		}
	}
	if err == nil {
		return true
	}

	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return false
	}

	writeError(w, http.StatusBadRequest, "Invalid request body")
	return false
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return 0, false
	}

	return id, true
}
