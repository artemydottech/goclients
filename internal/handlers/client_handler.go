package handlers

import (
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

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.CreateClient(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		log.Printf("CreateClient: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save client")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	var (
		clients []models.Client
		err     error
	)

	if raw := r.URL.Query().Get("company_id"); raw != "" {
		companyID, convErr := strconv.Atoi(raw)
		if convErr != nil {
			writeError(w, http.StatusBadRequest, "invalid company_id")
			return
		}
		clients, err = h.service.GetClientsByCompany(companyID)
	} else {
		clients, err = h.service.GetAllClients()
	}

	if err != nil {
		log.Printf("GetAllClients: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load clients")
		return
	}

	writeJSON(w, clients)
}

func (h *ClientHandler) GetClientById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	client, err := h.service.GetClientById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		log.Printf("GetClientById: %v", err)
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, client)
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = h.service.DeleteClientById(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		log.Printf("DeleteClient: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
