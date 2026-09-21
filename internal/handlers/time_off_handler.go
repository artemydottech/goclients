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

type TimeOffServ interface {
	CreateTimeOff(employeeID int, t models.TimeOff) (int64, error)
	GetTimeOffByEmployee(employeeID int) ([]models.TimeOff, error)
	DeleteTimeOffById(id int) error
}

type TimeOffHandler struct {
	service TimeOffServ
}

func NewTimeOffHandler(s TimeOffServ) *TimeOffHandler {
	return &TimeOffHandler{service: s}
}

func (h *TimeOffHandler) CreateTimeOff(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	var input models.TimeOff
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateTimeOff(employeeID, input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "employee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("CreateTimeOff: %v", err)
		http.Error(w, "failed to save time off", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *TimeOffHandler) GetTimeOff(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	periods, err := h.service.GetTimeOffByEmployee(employeeID)
	if err != nil {
		log.Printf("GetTimeOff: %v", err)
		http.Error(w, "failed to load time off", http.StatusInternalServerError)
		return
	}

	writeJSON(w, periods)
}

func (h *TimeOffHandler) DeleteTimeOff(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTimeOffById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "time off not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("DeleteTimeOff: %v", err)
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
