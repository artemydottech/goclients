package models

import (
	"testing"
	"time"
)

func TestAppointmentFilterValidation(t *testing.T) {
	at := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
	for _, f := range []AppointmentFilter{{}, {CompanyID: -1}, {CompanyID: 1, ClientID: -1}, {CompanyID: 1, EmployeeID: -1}, {CompanyID: 1, ServiceID: -1}, {CompanyID: 1, Status: "other"}, {CompanyID: 1, From: at, To: at}, {CompanyID: 1, Page: Page{Offset: -1}}} {
		if err := f.Validate(); err == nil {
			t.Fatalf("accepted %+v", f)
		}
	}
	for _, status := range []AppointmentStatus{"", AppointmentPending, AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow} {
		if err := (AppointmentFilter{CompanyID: 1, Status: status, From: at, To: at.Add(time.Hour)}).Validate(); err != nil {
			t.Fatal(err)
		}
	}
}
