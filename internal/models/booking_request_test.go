package models

import (
	"strings"
	"testing"
	"time"
)

func TestBookingFingerprintCanonicalInput(t *testing.T) {
	a := Appointment{ClientID: 1, EmployeeID: 2, ServiceID: 3, StartsAt: time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC)}
	first, err := BookingFingerprint(a)
	if err != nil {
		t.Fatal(err)
	}
	a.Status = AppointmentPending
	a.StartsAt = a.StartsAt.In(time.FixedZone("local", 18000))
	a.Price = 999
	a.ID = 42
	second, err := BookingFingerprint(a)
	if err != nil || first != second {
		t.Fatalf("canonical: %s %s %v", first, second, err)
	}
	a.Comment = "changed"
	third, _ := BookingFingerprint(a)
	if first == third {
		t.Fatal("comment ignored")
	}
	for _, key := range []string{"", "has space", "line\n", strings.Repeat("a", 129), "ключ"} {
		if ValidateBookingKey(key) == nil {
			t.Fatalf("accepted %q", key)
		}
	}
}
