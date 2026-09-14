package models

import "time"

type AppointmentStatus string

const (
	AppointmentPending   AppointmentStatus = "pending"
	AppointmentConfirmed AppointmentStatus = "confirmed"
	AppointmentCancelled AppointmentStatus = "cancelled"
	AppointmentCompleted AppointmentStatus = "completed"
	AppointmentNoShow    AppointmentStatus = "no_show"
)

type Appointment struct {
	ID         int               `json:"id"`
	CompanyID  int               `json:"company_id"`
	ClientID   int               `json:"client_id"`
	EmployeeID int               `json:"employee_id"`
	ServiceID  int               `json:"service_id"`
	StartsAt   time.Time         `json:"starts_at"`
	EndsAt     time.Time         `json:"ends_at"`
	Status     AppointmentStatus `json:"status"`
	Comment    string            `json:"comment"`

	Price float64 `json:"price"`
}

func (s AppointmentStatus) Valid() bool {
	switch s {
	case AppointmentPending, AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow:
		return true
	default:
		return false
	}
}

var statusTransitions = map[AppointmentStatus][]AppointmentStatus{
	AppointmentPending:   {AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow},
	AppointmentConfirmed: {AppointmentCancelled, AppointmentCompleted, AppointmentNoShow},
}

func (s AppointmentStatus) CanBecome(next AppointmentStatus) bool {
	if s == next {
		return true
	}

	for _, allowed := range statusTransitions[s] {
		if allowed == next {
			return true
		}
	}

	return false
}

func (s AppointmentStatus) Blocks() bool {
	return s != AppointmentCancelled
}

func (s AppointmentStatus) Active() bool {
	return s == AppointmentPending || s == AppointmentConfirmed
}

func (s AppointmentStatus) Happened() bool {
	return s == AppointmentCompleted || s == AppointmentNoShow
}
