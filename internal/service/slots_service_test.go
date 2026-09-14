package service

import (
	"errors"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type stubSchedule struct {
	day     models.WorkingDay
	working bool
}

func (s stubSchedule) WorkingDay(int, int) (models.WorkingDay, bool, error) {
	return s.day, s.working, nil
}

type stubCompanyLookup struct {
	company models.Company
}

func (l stubCompanyLookup) GetCompanyById(int) (models.Company, error) {
	return l.company, nil
}

var utcCompany = stubCompanyLookup{company: models.Company{ID: 10, Timezone: "UTC"}}

type stubTimeOff struct {
	periods []models.TimeOff
}

func (s stubTimeOff) GetTimeOffInRange(_ int, from, to time.Time) ([]models.TimeOff, error) {
	matching := []models.TimeOff{}
	for _, period := range s.periods {
		if from.Before(period.EndsAt) && to.After(period.StartsAt) {
			matching = append(matching, period)
		}
	}
	return matching, nil
}

var noTimeOff = stubTimeOff{}

type stubCalendar struct {
	booked []models.Appointment
}

func (c stubCalendar) GetAppointmentsByEmployee(int, time.Time, time.Time) ([]models.Appointment, error) {
	return c.booked, nil
}

var slotsDate = time.Date(2030, 3, 4, 0, 0, 0, 0, time.UTC)

func newSlotsService(schedule stubSchedule, calendar stubCalendar) *SlotsService {
	svc := NewSlotsService(
		schedule,
		calendar,
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 60}}},
		stubAssignments{performs: true},
		utcCompany,
		noTimeOff,
	)
	svc.now = func() time.Time { return slotsDate.AddDate(0, 0, -1) }

	return svc
}

func workingDay(from, to string) stubSchedule {
	return stubSchedule{
		day:     models.WorkingDay{Weekday: int(slotsDate.Weekday()), StartsAt: from, EndsAt: to},
		working: true,
	}
}

func TestFreeSlotsWalksTheWorkingDay(t *testing.T) {
	slots, err := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	want := []string{"10:00", "11:00", "12:00"}
	if len(slots) != len(want) {
		t.Fatalf("получено %d слотов, ожидалось %d: %v", len(slots), len(want), slots)
	}

	for i, expected := range want {
		if got := slots[i].Format("15:04"); got != expected {
			t.Errorf("слот %d = %s, ожидался %s", i, got, expected)
		}
	}
}

func TestFreeSlotsNeverRunPastTheEndOfTheDay(t *testing.T) {
	slots, err := newSlotsService(workingDay("10:00", "12:30"), stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 30)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	last := slots[len(slots)-1]
	if got := last.Format("15:04"); got != "11:30" {
		t.Errorf("последний слот %s, ожидался 11:30 — часовая услуга должна укладываться в день", got)
	}
}

func TestFreeSlotsSkipBookedTime(t *testing.T) {
	booked := []models.Appointment{{
		StartsAt: slotsDate.Add(11 * time.Hour),
		EndsAt:   slotsDate.Add(12 * time.Hour),
		Status:   models.AppointmentConfirmed,
	}}

	slots, err := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{booked: booked}).
		FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	for _, slot := range slots {
		if slot.Format("15:04") == "11:00" {
			t.Fatalf("занятое время попало в слоты: %v", slots)
		}
	}

	if len(slots) != 2 {
		t.Errorf("получено %d слотов, ожидалось 2: %v", len(slots), slots)
	}
}

func TestFreeSlotsIgnoreCancelledAppointments(t *testing.T) {
	booked := []models.Appointment{{
		StartsAt: slotsDate.Add(11 * time.Hour),
		EndsAt:   slotsDate.Add(12 * time.Hour),
		Status:   models.AppointmentCancelled,
	}}

	slots, err := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{booked: booked}).
		FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 3 {
		t.Errorf("отменённая запись не должна занимать слот, получено %v", slots)
	}
}

func TestFreeSlotsAreEmptyOnADayOff(t *testing.T) {
	slots, err := newSlotsService(stubSchedule{working: false}, stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 0 {
		t.Errorf("в выходной слотов быть не должно, получено %v", slots)
	}
}

func TestFreeSlotsDropPastTimes(t *testing.T) {
	svc := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{})
	svc.now = func() time.Time { return slotsDate.Add(11*time.Hour + 30*time.Minute) }

	slots, err := svc.FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 1 || slots[0].Format("15:04") != "12:00" {
		t.Errorf("ожидался только слот 12:00, получено %v", slots)
	}
}

func TestFreeSlotsRejectMasterWhoDoesNotPerformTheService(t *testing.T) {
	svc := NewSlotsService(
		workingDay("10:00", "13:00"),
		stubCalendar{},
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 60}}},
		stubAssignments{performs: false},
		utcCompany,
		noTimeOff,
	)

	_, err := svc.FreeSlots(2, 3, slotsDate, 60)

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestFreeSlotsFallBackToTheDefaultStep(t *testing.T) {
	slots, err := newSlotsService(workingDay("10:00", "11:15"), stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 0)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 2 {
		t.Fatalf("при шаге по умолчанию ожидалось 2 слота, получено %v", slots)
	}
	if slots[1].Format("15:04") != "10:15" {
		t.Errorf("второй слот %s, ожидался 10:15", slots[1].Format("15:04"))
	}
}

func TestFreeSlotsFollowTheCompanyTimezone(t *testing.T) {
	svc := newSlotsService(workingDay("10:00", "12:00"), stubCalendar{})
	svc.companies = stubCompanyLookup{company: models.Company{ID: 10, Timezone: "Asia/Yekaterinburg"}}

	slots, err := svc.FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 2 {
		t.Fatalf("получено %d слотов, ожидалось 2: %v", len(slots), slots)
	}

	want := time.Date(2030, 3, 4, 5, 0, 0, 0, time.UTC)
	if !slots[0].Equal(want) {
		t.Errorf("первый слот %v, ожидалось 10:00 по Екатеринбургу = %v", slots[0], want)
	}

	if got := slots[0].Format("15:04 -07:00"); got != "10:00 +05:00" {
		t.Errorf("слот отдан как %s, ожидалось местное время с поясом", got)
	}
}

func TestFreeSlotsSkipTimeOff(t *testing.T) {
	svc := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{})
	svc.timeOff = stubTimeOff{periods: []models.TimeOff{{
		StartsAt: slotsDate.Add(11 * time.Hour),
		EndsAt:   slotsDate.Add(12 * time.Hour),
		Reason:   "врач",
	}}}

	slots, err := svc.FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(slots) != 2 {
		t.Fatalf("получено %d слотов, ожидалось 2: %v", len(slots), slots)
	}
	for _, slot := range slots {
		if slot.Format("15:04") == "11:00" {
			t.Fatalf("слот внутри отгула: %v", slots)
		}
	}
}

func TestFreeSlotsAreEmptyDuringVacation(t *testing.T) {
	svc := newSlotsService(workingDay("10:00", "13:00"), stubCalendar{})
	svc.timeOff = stubTimeOff{periods: []models.TimeOff{{
		StartsAt: slotsDate.AddDate(0, 0, -3),
		EndsAt:   slotsDate.AddDate(0, 0, 10),
		Reason:   "отпуск",
	}}}

	slots, err := svc.FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("в отпуске слотов быть не должно, получено %v", slots)
	}
}

func workingDayWithBreak(from, to, breakFrom, breakTo string) stubSchedule {
	schedule := workingDay(from, to)
	schedule.day.BreakStartsAt = breakFrom
	schedule.day.BreakEndsAt = breakTo
	return schedule
}

func TestFreeSlotsSkipTheBreak(t *testing.T) {
	slots, err := newSlotsService(workingDayWithBreak("10:00", "14:00", "12:00", "13:00"), stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 60)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	got := make([]string, 0, len(slots))
	for _, slot := range slots {
		got = append(got, slot.Format("15:04"))
	}

	want := []string{"10:00", "11:00", "13:00"}
	if len(got) != len(want) {
		t.Fatalf("слоты %v, ожидались %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("слоты %v, ожидались %v", got, want)
			break
		}
	}
}

func TestFreeSlotsDoNotStraddleTheBreak(t *testing.T) {
	slots, err := newSlotsService(workingDayWithBreak("10:00", "14:00", "12:00", "13:00"), stubCalendar{}).
		FreeSlots(2, 3, slotsDate, 30)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	for _, slot := range slots {
		if slot.Format("15:04") == "11:30" {
			t.Fatalf("часовая услуга с 11:30 залезает на перерыв: %v", slots)
		}
	}
}
