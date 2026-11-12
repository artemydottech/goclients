package models

import "time"

type AppointmentEvent struct {
	ID            int64        `json:"id"`
	AppointmentID int          `json:"appointment_id"`
	Kind          string       `json:"kind"`
	OccurredAt    time.Time    `json:"occurred_at"`
	Before        *Appointment `json:"before"`
	After         Appointment  `json:"after"`
}
