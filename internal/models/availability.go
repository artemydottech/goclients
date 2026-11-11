package models

import "time"

type AvailabilityQuery struct {
	ServiceID, EmployeeID, Days, Step, Limit int
	Date                                     time.Time
}

func (q AvailabilityQuery) Validate() error {
	if q.ServiceID <= 0 || q.EmployeeID < 0 {
		return Invalid("invalid service or employee id")
	}
	if q.Date.IsZero() {
		return Invalid("date is required")
	}
	if q.Days < 1 || q.Days > 14 {
		return Invalid("days must be between 1 and 14")
	}
	if q.Step < 1 || q.Step > 1440 {
		return Invalid("step must be between 1 and 1440")
	}
	if q.Limit < 1 || q.Limit > 200 {
		return Invalid("limit must be between 1 and 200")
	}
	return nil
}

type OpenSlot struct {
	EmployeeID   int       `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
}
