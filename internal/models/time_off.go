package models

import "time"

// TimeOff — время, когда мастер не принимает: отпуск, больничный, отгул.
// В отличие от выходного по графику это разовый интервал с конкретными датами.
type TimeOff struct {
	ID         int       `json:"id"`
	EmployeeID int       `json:"employee_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	Reason     string    `json:"reason"`
}
