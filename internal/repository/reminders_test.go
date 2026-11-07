package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type reminderFixture struct {
	db        *sql.DB
	repo      *AppointmentRepository
	companyID int
	clientID  int
	employee  int
	service   int
}

func newReminderFixture(t *testing.T) reminderFixture {
	t.Helper()

	db := newTestDB(t)
	ctx := context.Background()

	companyID, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: "Romashka", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	clientID, err := NewClientRepository(db).Create(ctx, models.Client{CompanyID: int(companyID), Name: "Ivan", Phone: "79990000000"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	employeeID, err := NewEmployeeRepository(db).Create(ctx, models.Employee{CompanyID: int(companyID), Name: "Anna", Surname: "Ivanova"})
	if err != nil {
		t.Fatalf("employee: %v", err)
	}
	serviceID, err := NewServiceRepository(db).Create(ctx, models.Service{CompanyID: int(companyID), Name: "Haircut", Duration: 60})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	return reminderFixture{
		db: db, repo: NewAppointmentRepository(db),
		companyID: int(companyID), clientID: int(clientID), employee: int(employeeID), service: int(serviceID),
	}
}

func (f reminderFixture) book(t *testing.T, start time.Time, status models.AppointmentStatus) int {
	t.Helper()

	id, err := f.repo.Create(context.Background(), models.Appointment{
		CompanyID: f.companyID, ClientID: f.clientID, EmployeeID: f.employee, ServiceID: f.service,
		StartsAt: start, EndsAt: start.Add(time.Hour), Status: status,
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	return int(id)
}

func TestDueRemindersReturnTheClientContacts(t *testing.T) {
	f := newReminderFixture(t)
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	id := f.book(t, now.Add(20*time.Hour), models.AppointmentConfirmed)

	got, err := f.repo.GetDueReminders(context.Background(), now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetDueReminders: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d reminders, want 1", len(got))
	}
	if got[0].Appointment.ID != id || got[0].ClientName != "Ivan" || got[0].ClientPhone != "79990000000" {
		t.Errorf("got %+v", got[0])
	}
}

func TestDueRemindersRespectTheWindowAndStatus(t *testing.T) {
	f := newReminderFixture(t)
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)

	f.book(t, now.Add(-time.Hour), models.AppointmentConfirmed)
	f.book(t, now.Add(30*time.Hour), models.AppointmentConfirmed)
	f.book(t, now.Add(5*time.Hour), models.AppointmentCancelled)
	f.book(t, now.Add(6*time.Hour), models.AppointmentNoShow)
	inside := f.book(t, now.Add(7*time.Hour), models.AppointmentPending)
	edge := f.book(t, now.Add(24*time.Hour), models.AppointmentPending)

	got, err := f.repo.GetDueReminders(context.Background(), now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetDueReminders: %v", err)
	}

	var gotIDs []int
	for _, reminder := range got {
		gotIDs = append(gotIDs, reminder.Appointment.ID)
	}
	if !equalInts(gotIDs, []int{inside, edge}) {
		t.Errorf("ids = %v, want [%d %d]", gotIDs, inside, edge)
	}
}

func TestMarkedAppointmentsAreNotRemindedTwice(t *testing.T) {
	f := newReminderFixture(t)
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	id := f.book(t, now.Add(5*time.Hour), models.AppointmentConfirmed)
	ctx := context.Background()

	if err := f.repo.MarkReminded(ctx, id, now); err != nil {
		t.Fatalf("MarkReminded: %v", err)
	}

	got, err := f.repo.GetDueReminders(ctx, now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetDueReminders: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d reminders after marking, want 0", len(got))
	}
}

func TestMarkRemindedMissingAppointment(t *testing.T) {
	f := newReminderFixture(t)

	err := f.repo.MarkReminded(context.Background(), 404, time.Now())

	if err == nil {
		t.Fatal("expected an error for a missing appointment")
	}
}
func TestReschedulingAllowsANewReminder(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	id := f.book(t, now.Add(time.Hour), models.AppointmentPending)
	if err := f.repo.MarkReminded(ctx, id, now); err != nil {
		t.Fatal(err)
	}
	a, err := f.repo.GetAppointmentById(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	a.StartsAt = now.Add(2 * time.Hour)
	a.EndsAt = a.StartsAt.Add(time.Hour)
	if busy, err := f.repo.MoveIfFree(ctx, id, a); err != nil || busy {
		t.Fatalf("move: busy=%v %v", busy, err)
	}
	due, err := f.repo.GetDueReminders(ctx, now, now.Add(24*time.Hour))
	if err != nil || len(due) != 1 || due[0].Appointment.ID != id {
		t.Fatalf("due: %+v %v", due, err)
	}
}
