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

type AppointmentServ interface {
	Book(a models.Appointment) (int64, error)
	GetAllAppointments() ([]models.Appointment, error)
	GetAppointmentsByEmployee(employeeID int, from, to time.Time) ([]models.Appointment, error)
	GetAppointmentsByClient(clientID int) ([]models.Appointment, error)
	GetAppointmentById(id int) (models.Appointment, error)
	SetStatus(id int, status models.AppointmentStatus) error
	Reschedule(id int, startsAt time.Time, employeeID int) (models.Appointment, error)
	DeleteAppointmentById(id int) error
}

type AppointmentHandler struct {
	service AppointmentServ
}

func NewAppointmentHandler(s AppointmentServ) *AppointmentHandler {
	return &AppointmentHandler{service: s}
}

func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	var input models.Appointment

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.Book(input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("CreateAppointment: %v", err)
		http.Error(w, "Ошибка сохранения записи", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *AppointmentHandler) GetAllAppointments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if raw := query.Get("client_id"); raw != "" {
		clientID, err := strconv.Atoi(raw)
		if err != nil {
			http.Error(w, "Неправильный client_id", http.StatusBadRequest)
			return
		}

		appointments, err := h.service.GetAppointmentsByClient(clientID)
		if err != nil {
			log.Printf("GetAppointmentsByClient: %v", err)
			http.Error(w, "Ошибка запроса записей", http.StatusInternalServerError)
			return
		}

		writeJSON(w, appointments)
		return
	}

	if raw := query.Get("employee_id"); raw != "" {
		employeeID, err := strconv.Atoi(raw)
		if err != nil {
			http.Error(w, "Неправильный employee_id", http.StatusBadRequest)
			return
		}

		from, to, err := parseRange(query.Get("from"), query.Get("to"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		appointments, err := h.service.GetAppointmentsByEmployee(employeeID, from, to)
		if err != nil {
			log.Printf("GetAppointmentsByEmployee: %v", err)
			http.Error(w, "Ошибка запроса записей", http.StatusInternalServerError)
			return
		}

		writeJSON(w, appointments)
		return
	}

	appointments, err := h.service.GetAllAppointments()
	if err != nil {
		log.Printf("GetAllAppointments: %v", err)
		http.Error(w, "Ошибка запроса записей", http.StatusInternalServerError)
		return
	}

	writeJSON(w, appointments)
}

// parseRange читает окно календаря. Без параметров берётся ближайшая неделя —
// иначе выборка по мастеру вернула бы всю его историю.
func parseRange(rawFrom, rawTo string) (time.Time, time.Time, error) {
	from := time.Now().UTC()
	to := from.AddDate(0, 0, 7)

	if rawFrom != "" {
		parsed, err := time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("Параметр from должен быть в формате RFC3339")
		}
		from = parsed.UTC()
		to = from.AddDate(0, 0, 7)
	}

	if rawTo != "" {
		parsed, err := time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("Параметр to должен быть в формате RFC3339")
		}
		to = parsed.UTC()
	}

	if !to.After(from) {
		return time.Time{}, time.Time{}, errors.New("Конец периода должен быть позже начала")
	}

	return from, to, nil
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

func (h *AppointmentHandler) GetAppointmentById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	appointment, err := h.service.GetAppointmentById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("GetAppointmentById: %v", err)
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		return
	}

	writeJSON(w, appointment)
}

type statusInput struct {
	Status models.AppointmentStatus `json:"status"`
}

func (h *AppointmentHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var input statusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.SetStatus(id, input.Status)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("SetStatus: %v", err)
		http.Error(w, "Ошибка смены статуса", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppointmentHandler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteAppointmentById(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("DeleteAppointment: %v", err)
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type rescheduleInput struct {
	StartsAt   time.Time `json:"starts_at"`
	EmployeeID int       `json:"employee_id"`
}

func (h *AppointmentHandler) Reschedule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var input rescheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	appointment, err := h.service.Reschedule(id, input.StartsAt, input.EmployeeID)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Reschedule: %v", err)
		http.Error(w, "Ошибка переноса записи", http.StatusInternalServerError)
		return
	}

	writeJSON(w, appointment)
}
