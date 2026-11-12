package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type AvailabilityAssignments interface {
	GetEmployeesByService(context.Context, int) ([]models.Employee, error)
}

type AvailabilitySlots interface {
	FreeSlots(context.Context, int, int, time.Time, int) ([]time.Time, error)
}

type AvailabilityService struct {
	services    ServiceLookup
	assignments AvailabilityAssignments
	slots       AvailabilitySlots
}

func NewAvailabilityService(s ServiceLookup, a AvailabilityAssignments, slots AvailabilitySlots) *AvailabilityService {
	return &AvailabilityService{services: s, assignments: a, slots: slots}
}

func (s *AvailabilityService) Search(ctx context.Context, q models.AvailabilityQuery) ([]models.OpenSlot, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	item, err := s.services.GetServiceById(ctx, q.ServiceID)
	if err != nil {
		return nil, err
	}
	employees, err := s.assignments.GetEmployeesByService(ctx, q.ServiceID)
	if err != nil {
		return nil, err
	}
	eligible := []models.Employee{}
	for _, e := range employees {
		if e.CompanyID == item.CompanyID && (q.EmployeeID == 0 || q.EmployeeID == e.ID) {
			eligible = append(eligible, e)
		}
	}
	if q.EmployeeID > 0 && len(eligible) == 0 {
		return nil, models.Invalid("employee does not provide this service")
	}
	result := []models.OpenSlot{}
	for day := 0; day < q.Days; day++ {
		daily := []models.OpenSlot{}
		for _, e := range eligible {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			slots, err := s.slots.FreeSlots(ctx, e.ID, q.ServiceID, q.Date.AddDate(0, 0, day), q.Step)
			if err != nil {
				return nil, err
			}
			for _, start := range slots {
				daily = append(daily, models.OpenSlot{EmployeeID: e.ID, EmployeeName: strings.TrimSpace(e.Name + " " + e.Surname), StartsAt: start, EndsAt: start.Add(time.Duration(item.Duration) * time.Minute)})
			}
		}
		sort.Slice(daily, func(i, j int) bool {
			if daily[i].StartsAt.Equal(daily[j].StartsAt) {
				return daily[i].EmployeeID < daily[j].EmployeeID
			}
			return daily[i].StartsAt.Before(daily[j].StartsAt)
		})
		remaining := q.Limit - len(result)
		if len(daily) > remaining {
			daily = daily[:remaining]
		}
		result = append(result, daily...)
		if len(result) == q.Limit {
			break
		}
	}
	return result, nil
}
