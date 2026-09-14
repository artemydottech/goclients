package models

import "time"

type TimeOff struct {
	ID         int       `json:"id"`
	EmployeeID int       `json:"employee_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	Reason     string    `json:"reason"`
}
