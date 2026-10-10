package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type ScheduleServ interface {
	SetEmployeeSchedule(employeeID int, days []models.WorkingDay) error
	GetEmployeeSchedule(employeeID int) ([]models.WorkingDay, error)
}

type SlotsServ interface {
	FreeSlots(employeeID, serviceID int, date time.Time, step int) ([]time.Time, error)
}

type ScheduleHandler struct {
	schedule ScheduleServ
	slots    SlotsServ
}

func NewScheduleHandler(schedule ScheduleServ, slots SlotsServ) *ScheduleHandler {
	return &ScheduleHandler{schedule: schedule, slots: slots}
}

type scheduleInput struct {
	Days []models.WorkingDay `json:"days"`
}

func (h *ScheduleHandler) SetEmployeeSchedule(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var input scheduleInput
	if !decodeJSON(w, r, &input) {
		return
	}

	err = h.schedule.SetEmployeeSchedule(employeeID, input.Days)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		log.Printf("SetEmployeeSchedule: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save schedule")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) GetEmployeeSchedule(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	days, err := h.schedule.GetEmployeeSchedule(employeeID)
	if err != nil {
		log.Printf("GetEmployeeSchedule: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load schedule")
		return
	}

	writeJSON(w, days)
}

func (h *ScheduleHandler) GetFreeSlots(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	employeeID, err := strconv.Atoi(query.Get("employee_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "employee_id is required")
		return
	}

	serviceID, err := strconv.Atoi(query.Get("service_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "service_id is required")
		return
	}

	date, err := time.Parse("2006-01-02", query.Get("date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "date is required in YYYY-MM-DD format")
		return
	}

	step := 0
	if raw := query.Get("step"); raw != "" {
		step, err = strconv.Atoi(raw)
		if err != nil || step <= 0 {
			writeError(w, http.StatusBadRequest, "step must be a positive number of minutes")
			return
		}
	}

	slots, err := h.slots.FreeSlots(employeeID, serviceID, date, step)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		log.Printf("GetFreeSlots: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to compute slots")
		return
	}

	writeJSON(w, slots)
}
