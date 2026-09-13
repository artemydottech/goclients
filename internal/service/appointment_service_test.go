package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

type stubAppointmentRepo struct {
	created models.Appointment
	calls   int
	overlap bool
}

func (r *stubAppointmentRepo) Create(a models.Appointment) (int64, error) {
	r.calls++
	r.created = a
	return 11, nil
}

func (r *stubAppointmentRepo) GetAllAppointments() ([]models.Appointment, error) {
	return nil, nil
}

func (r *stubAppointmentRepo) GetAppointmentsByEmployee(int, time.Time, time.Time) ([]models.Appointment, error) {
	return nil, nil
}

func (r *stubAppointmentRepo) GetAppointmentsByClient(int) ([]models.Appointment, error) {
	return nil, nil
}

func (r *stubAppointmentRepo) GetAppointmentById(int) (models.Appointment, error) {
	return models.Appointment{}, nil
}

func (r *stubAppointmentRepo) CreateIfFree(a models.Appointment) (int64, bool, error) {
	if r.overlap {
		return 0, true, nil
	}
	id, err := r.Create(a)
	return id, false, err
}

func (r *stubAppointmentRepo) UpdateStatus(int, models.AppointmentStatus) error { return nil }

func (r *stubAppointmentRepo) DeleteAppointmentById(int) error { return nil }

type stubClientLookup struct {
	client models.Client
	err    error
}

func (l stubClientLookup) GetClientById(int) (models.Client, error) {
	return l.client, l.err
}

type stubAssignments struct {
	performs bool
}

func (a stubAssignments) EmployeePerformsService(int, int) (bool, error) {
	return a.performs, nil
}

func newAppointmentService(repo *stubAppointmentRepo, opts ...func(*AppointmentService)) *AppointmentService {
	svc := NewAppointmentService(
		repo,
		stubClientLookup{client: models.Client{ID: 1, CompanyID: 10}},
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{
			3: {ID: 3, CompanyID: 10, Duration: 90},
		}},
		stubAssignments{performs: true},
		workingDay("10:00", "20:00"),
		utcCompany,
	)
	svc.now = func() time.Time { return slotsDate.AddDate(0, 0, -1) }

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func futureBooking() models.Appointment {
	return models.Appointment{
		ClientID:   1,
		EmployeeID: 2,
		ServiceID:  3,
		StartsAt:   slotsDate.Add(12 * time.Hour),
	}
}

func TestBookComputesEndFromServiceDuration(t *testing.T) {
	repo := &stubAppointmentRepo{}

	booking := futureBooking()
	booking.EndsAt = booking.StartsAt.Add(5 * time.Minute)

	if _, err := newAppointmentService(repo).Book(booking); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	got := repo.created.EndsAt.Sub(repo.created.StartsAt)
	if got != 90*time.Minute {
		t.Errorf("длительность записи %v, ожидалось 90m — конец из тела запроса игнорируется", got)
	}
}

func TestBookDefaultsToPending(t *testing.T) {
	repo := &stubAppointmentRepo{}

	if _, err := newAppointmentService(repo).Book(futureBooking()); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if repo.created.Status != models.AppointmentPending {
		t.Errorf("статус %q, ожидался pending", repo.created.Status)
	}
}

func TestBookStampsTheCompanyFromTheEmployee(t *testing.T) {
	repo := &stubAppointmentRepo{}

	booking := futureBooking()
	booking.CompanyID = 999

	if _, err := newAppointmentService(repo).Book(booking); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if repo.created.CompanyID != 10 {
		t.Errorf("company_id %d, ожидался 10 — компания берётся у сотрудника", repo.created.CompanyID)
	}
}

func TestBookRejectsPastAndZeroTime(t *testing.T) {
	cases := map[string]time.Time{
		"прошлое":      time.Now().Add(-time.Hour),
		"нулевое":      {},
		"ровно сейчас": time.Now().Add(-time.Millisecond),
	}

	for name, startsAt := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubAppointmentRepo{}

			booking := futureBooking()
			booking.StartsAt = startsAt

			_, err := newAppointmentService(repo).Book(booking)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}

			if repo.calls != 0 {
				t.Error("запись не должна сохраняться")
			}
		})
	}
}

func TestBookRejectsBusyEmployee(t *testing.T) {
	repo := &stubAppointmentRepo{overlap: true}

	_, err := newAppointmentService(repo).Book(futureBooking())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.calls != 0 {
		t.Error("пересекающаяся запись не должна сохраняться")
	}
}

func TestBookRejectsEmployeeWhoDoesNotPerformTheService(t *testing.T) {
	repo := &stubAppointmentRepo{}

	svc := NewAppointmentService(
		repo,
		stubClientLookup{client: models.Client{ID: 1, CompanyID: 10}},
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 90}}},
		stubAssignments{performs: false},
		workingDay("10:00", "20:00"),
		utcCompany,
	)

	_, err := svc.Book(futureBooking())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestBookRejectsCrossCompanyParts(t *testing.T) {
	repo := &stubAppointmentRepo{}

	svc := NewAppointmentService(
		repo,
		stubClientLookup{client: models.Client{ID: 1, CompanyID: 99}},
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 90}}},
		stubAssignments{performs: true},
		workingDay("10:00", "20:00"),
		utcCompany,
	)

	_, err := svc.Book(futureBooking())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.calls != 0 {
		t.Error("запись из разных компаний не должна сохраняться")
	}
}

func TestBookReportsMissingClient(t *testing.T) {
	repo := &stubAppointmentRepo{}

	svc := NewAppointmentService(
		repo,
		stubClientLookup{err: sql.ErrNoRows},
		stubEmployeeLookup{employee: models.Employee{ID: 2, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{3: {ID: 3, CompanyID: 10, Duration: 90}}},
		stubAssignments{performs: true},
		workingDay("10:00", "20:00"),
		utcCompany,
	)

	_, err := svc.Book(futureBooking())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestBookRejectsUnknownStatus(t *testing.T) {
	repo := &stubAppointmentRepo{}

	booking := futureBooking()
	booking.Status = "вроде бы"

	_, err := newAppointmentService(repo).Book(booking)

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestSetStatusRejectsUnknownStatus(t *testing.T) {
	err := newAppointmentService(&stubAppointmentRepo{}).SetStatus(1, "готово")

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestCancelledAppointmentsDoNotBlockTheSlot(t *testing.T) {
	if models.AppointmentCancelled.Blocks() {
		t.Error("отменённая запись не должна занимать время мастера")
	}

	for _, status := range []models.AppointmentStatus{
		models.AppointmentPending,
		models.AppointmentConfirmed,
		models.AppointmentCompleted,
	} {
		if !status.Blocks() {
			t.Errorf("статус %q должен занимать время мастера", status)
		}
	}
}

func TestBookRejectsTimeOutsideWorkingHours(t *testing.T) {
	cases := map[string]time.Time{
		"до открытия":             slotsDate.Add(9 * time.Hour),
		"не успевает до закрытия": slotsDate.Add(19*time.Hour + 30*time.Minute),
		"ночью": slotsDate.Add(3 * time.Hour),
	}

	for name, startsAt := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubAppointmentRepo{}

			booking := futureBooking()
			booking.StartsAt = startsAt

			_, err := newAppointmentService(repo).Book(booking)

			var validationErr models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("ожидалась ValidationError, получено %v", err)
			}

			if repo.calls != 0 {
				t.Error("запись вне рабочего времени не должна сохраняться")
			}
		})
	}
}

func TestBookAcceptsTheLastSlotThatFits(t *testing.T) {
	repo := &stubAppointmentRepo{}

	booking := futureBooking()
	booking.StartsAt = slotsDate.Add(18*time.Hour + 30*time.Minute)

	if _, err := newAppointmentService(repo).Book(booking); err != nil {
		t.Fatalf("90-минутная услуга в 18:30 заканчивается ровно в 20:00, получено %v", err)
	}
}

func TestBookRejectsDayOff(t *testing.T) {
	repo := &stubAppointmentRepo{}

	svc := newAppointmentService(repo, func(s *AppointmentService) {
		s.schedule = stubSchedule{working: false}
	})

	_, err := svc.Book(futureBooking())

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}
}

func TestBookReadsWorkingHoursInTheCompanyTimezone(t *testing.T) {
	cases := map[string]struct {
		startsAt time.Time
		ok       bool
	}{
		"10:00 по Екатеринбургу":               {slotsDate.Add(5 * time.Hour), true},
		"20:30 по Екатеринбургу, 15:30 по UTC": {slotsDate.Add(15*time.Hour + 30*time.Minute), false},
		"09:00 по Екатеринбургу":               {slotsDate.Add(4 * time.Hour), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &stubAppointmentRepo{}
			svc := newAppointmentService(repo, func(s *AppointmentService) {
				s.companies = stubCompanyLookup{company: models.Company{ID: 10, Timezone: "Asia/Yekaterinburg"}}
			})

			booking := futureBooking()
			booking.StartsAt = tc.startsAt

			_, err := svc.Book(booking)
			if tc.ok && err != nil {
				t.Fatalf("запись должна пройти, получено %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("запись вне рабочего времени по местному поясу прошла")
			}
		})
	}
}
