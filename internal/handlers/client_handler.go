package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientServ interface {
	CreateClient(c models.Client) (int64, error)
	GetAllClients() ([]models.Client, error)
	GetClientsByCompany(companyID int) ([]models.Client, error)
	GetClientById(id int) (models.Client, error)
	DeleteClientById(id int) error
}

type ClientHandler struct {
	service ClientServ
}

func NewClientHandler(s ClientServ) *ClientHandler {
	return &ClientHandler{service: s}
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var input models.Client

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateClient(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("CreateClient: %v", err)
		http.Error(w, "failed to save client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	var (
		clients []models.Client
		err     error
	)

	if raw := r.URL.Query().Get("company_id"); raw != "" {
		companyID, convErr := strconv.Atoi(raw)
		if convErr != nil {
			http.Error(w, "invalid company_id", http.StatusBadRequest)
			return
		}
		clients, err = h.service.GetClientsByCompany(companyID)
	} else {
		clients, err = h.service.GetAllClients()
	}

	if err != nil {
		log.Printf("GetAllClients: %v", err)
		http.Error(w, "failed to load clients", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clients)
}

func (h *ClientHandler) GetClientById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	client, err := h.service.GetClientById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("GetClientById: %v", err)
		http.Error(w, "request failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(client)
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteClientById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("DeleteClient: %v", err)
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
