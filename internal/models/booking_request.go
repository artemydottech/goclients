package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type BookingResult struct {
	ID       int64
	Replayed bool
}

type BookingRequest struct {
	Fingerprint   string
	AppointmentID int64
}

func ValidateBookingKey(key string) error {
	if len(key) < 1 || len(key) > 128 || strings.IndexFunc(key, func(r rune) bool { return r < 33 || r > 126 }) >= 0 {
		return Invalid("Idempotency-Key must contain 1 to 128 visible ASCII characters")
	}
	return nil
}

func BookingFingerprint(a Appointment) (string, error) {
	status := a.Status
	if status == "" {
		status = AppointmentPending
	}
	request := struct {
		ClientID, EmployeeID, ServiceID int
		StartsAt                        time.Time
		Comment                         string
		Status                          AppointmentStatus
	}{a.ClientID, a.EmployeeID, a.ServiceID, a.StartsAt.UTC(), a.Comment, status}
	raw, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw)), nil
}
