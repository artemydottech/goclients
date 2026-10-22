package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientServ interface {
	CreateClient(ctx context.Context, c models.Client) (int64, error)
	GetAllClients(ctx context.Context, page models.Page) ([]models.Client, error)
	GetClientsByCompany(ctx context.Context, companyID int, page models.Page) ([]models.Client, error)
	GetClientById(ctx context.Context, id int) (models.Client, error)
	UpdateClient(ctx context.Context, id int, c models.Client) error
	DeleteClientById(ctx context.Context, id int) error
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

	id, err := h.service.CreateClient(r.Context(), input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "CreateClient failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save client")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePage(w, r)
	if !ok {
		return
	}

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
		clients, err = h.service.GetClientsByCompany(r.Context(), companyID, page)
	} else {
		clients, err = h.service.GetAllClients(r.Context(), page)
	}

	if err != nil {
		slog.ErrorContext(r.Context(), "GetAllClients failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load clients")
		return
	}

	writeJSON(w, clients)
}

func (h *ClientHandler) GetClientById(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	client, err := h.service.GetClientById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "GetClientById failed", "err", err)
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, client)
}

func (h *ClientHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	var input models.Client
	if !decodeJSON(w, r, &input) {
		return
	}

	err := h.service.UpdateClient(r.Context(), id, input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "UpdateClient failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update client")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	err := h.service.DeleteClientById(r.Context(), id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "DeleteClient failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
