package models

import (
	"testing"
	"time"
)

func TestAvailabilityQueryBounds(t *testing.T) {
	valid := AvailabilityQuery{ServiceID: 1, Date: time.Now(), Days: 7, Step: 15, Limit: 50}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*AvailabilityQuery){func(q *AvailabilityQuery) { q.ServiceID = 0 }, func(q *AvailabilityQuery) { q.EmployeeID = -1 }, func(q *AvailabilityQuery) { q.Date = time.Time{} }, func(q *AvailabilityQuery) { q.Days = 15 }, func(q *AvailabilityQuery) { q.Days = 0 }, func(q *AvailabilityQuery) { q.Step = 1441 }, func(q *AvailabilityQuery) { q.Step = 0 }, func(q *AvailabilityQuery) { q.Limit = 201 }} {
		q := valid
		change(&q)
		if q.Validate() == nil {
			t.Fatalf("accepted %+v", q)
		}
	}
}
