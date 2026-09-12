package models

import "time"

type AppointmentStatus string

const (
	AppointmentPending   AppointmentStatus = "pending"
	AppointmentConfirmed AppointmentStatus = "confirmed"
	AppointmentCancelled AppointmentStatus = "cancelled"
	AppointmentCompleted AppointmentStatus = "completed"
)

// Appointment — запись клиента к мастеру на услугу. EndsAt не приходит
// снаружи: его считает сервис из длительности услуги.
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
}

func (s AppointmentStatus) Valid() bool {
	switch s {
	case AppointmentPending, AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted:
		return true
	default:
		return false
	}
}

// Blocks сообщает, занимает ли запись время мастера. Отменённая — не занимает,
// иначе освободившийся слот никто не смог бы забрать.
func (s AppointmentStatus) Blocks() bool {
	return s != AppointmentCancelled
}
