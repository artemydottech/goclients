package handlers

import (
	"database/sql"
	"encoding/json"
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
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var input scheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.schedule.SetEmployeeSchedule(employeeID, input.Days)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("SetEmployeeSchedule: %v", err)
		http.Error(w, "Ошибка сохранения графика", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) GetEmployeeSchedule(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	days, err := h.schedule.GetEmployeeSchedule(employeeID)
	if err != nil {
		log.Printf("GetEmployeeSchedule: %v", err)
		http.Error(w, "Ошибка запроса графика", http.StatusInternalServerError)
		return
	}

	writeJSON(w, days)
}

func (h *ScheduleHandler) GetFreeSlots(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	employeeID, err := strconv.Atoi(query.Get("employee_id"))
	if err != nil {
		http.Error(w, "Нужен employee_id", http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.Atoi(query.Get("service_id"))
	if err != nil {
		http.Error(w, "Нужен service_id", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", query.Get("date"))
	if err != nil {
		http.Error(w, "Нужна дата в формате ГГГГ-ММ-ДД", http.StatusBadRequest)
		return
	}

	step := 0
	if raw := query.Get("step"); raw != "" {
		step, err = strconv.Atoi(raw)
		if err != nil || step <= 0 {
			http.Error(w, "Шаг сетки должен быть положительным числом минут", http.StatusBadRequest)
			return
		}
	}

	slots, err := h.slots.FreeSlots(employeeID, serviceID, date, step)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("GetFreeSlots: %v", err)
		http.Error(w, "Ошибка расчёта слотов", http.StatusInternalServerError)
		return
	}

	writeJSON(w, slots)
}
