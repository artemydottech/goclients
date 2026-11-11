package service

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
	"time"
)

type availabilityAssignmentStub struct {
	employees []models.Employee
	err       error
}

func (a availabilityAssignmentStub) GetEmployeesByService(context.Context, int) ([]models.Employee, error) {
	return a.employees, a.err
}

type availabilitySlotStub struct {
	times map[int][]time.Time
	dates []time.Time
	err   error
}

func (s *availabilitySlotStub) FreeSlots(ctx context.Context, id, service int, date time.Time, step int) ([]time.Time, error) {
	s.dates = append(s.dates, date)
	return s.times[id], s.err
}
func availabilityFixture() (*AvailabilityService, *availabilitySlotStub, models.AvailabilityQuery) {
	at := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	slots := &availabilitySlotStub{times: map[int][]time.Time{2: {at.Add(9 * time.Hour), at.Add(11 * time.Hour)}, 1: {at.Add(10 * time.Hour)}}}
	svc := NewAvailabilityService(stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 45}}}, availabilityAssignmentStub{employees: []models.Employee{{ID: 2, CompanyID: 10, Name: "B"}, {ID: 1, CompanyID: 10, Name: "A"}, {ID: 9, CompanyID: 99, Name: "Foreign"}}}, slots)
	return svc, slots, models.AvailabilityQuery{ServiceID: 3, Date: at, Days: 1, Step: 15, Limit: 2}
}
func TestAvailabilityMergesChronologicallyAndCapsResults(t *testing.T) {
	svc, _, q := availabilityFixture()
	got, err := svc.Search(context.Background(), q)
	if err != nil || len(got) != 2 || got[0].EmployeeID != 2 || got[1].EmployeeID != 1 || got[0].EndsAt.Sub(got[0].StartsAt) != 45*time.Minute {
		t.Fatalf("availability: %+v %v", got, err)
	}
	q.EmployeeID = 1
	got, err = svc.Search(context.Background(), q)
	if err != nil || len(got) != 1 || got[0].EmployeeID != 1 {
		t.Fatalf("employee: %+v %v", got, err)
	}
}
