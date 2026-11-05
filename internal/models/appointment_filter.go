package models

import "time"

type AppointmentFilter struct {
	CompanyID  int
	ClientID   int
	EmployeeID int
	ServiceID  int
	Status     AppointmentStatus
	From       time.Time
	To         time.Time
	Page       Page
}

func (f AppointmentFilter) Validate() error {
	if f.CompanyID <= 0 {
		return Invalid("company_id must be positive")
	}
	if f.ClientID < 0 || f.EmployeeID < 0 || f.ServiceID < 0 {
		return Invalid("filter identifiers cannot be negative")
	}
	if f.Status != "" && !f.Status.Valid() {
		return Invalid("invalid appointment status")
	}
	if !f.From.IsZero() && !f.To.IsZero() && !f.To.After(f.From) {
		return Invalid("to must be after from")
	}
	return f.Page.Validate()
}
