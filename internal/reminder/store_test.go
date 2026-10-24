package reminder

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
	"github.com/artemydottech/goclients/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestWorkerWithARealDatabaseRemindsOnlyOnce(t *testing.T) {
	ctx := context.Background()
	db, err := repository.Open(ctx, filepath.Join(t.TempDir(), "reminders.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := repository.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	companyID, err := repository.NewCompanyRepository(db).Create(ctx, models.Company{Name: "Romashka", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	clientID, err := repository.NewClientRepository(db).Create(ctx, models.Client{CompanyID: int(companyID), Name: "Ivan", Phone: "79990000000"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	employeeID, err := repository.NewEmployeeRepository(db).Create(ctx, models.Employee{CompanyID: int(companyID), Name: "Anna", Surname: "Ivanova"})
	if err != nil {
		t.Fatalf("employee: %v", err)
	}
	serviceID, err := repository.NewServiceRepository(db).Create(ctx, models.Service{CompanyID: int(companyID), Name: "Haircut", Duration: 60})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	appointments := repository.NewAppointmentRepository(db)
	book := func(start time.Time) int {
		id, err := appointments.Create(ctx, models.Appointment{
			CompanyID: int(companyID), ClientID: int(clientID), EmployeeID: int(employeeID), ServiceID: int(serviceID),
			StartsAt: start, EndsAt: start.Add(time.Hour), Status: models.AppointmentConfirmed,
		})
		if err != nil {
			t.Fatalf("book: %v", err)
		}

		return int(id)
	}
	soon := book(fixedNow.Add(3 * time.Hour))
	book(fixedNow.Add(72 * time.Hour))

	notifier := &fakeNotifier{}
	worker := newTestWorker(appointments, notifier)

	sent, err := worker.RunOnce(ctx)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if sent != 1 || len(notifier.sent) != 1 || notifier.sent[0] != soon {
		t.Fatalf("first run sent=%d notified=%v, want only appointment %d", sent, notifier.sent, soon)
	}

	sent, err = worker.RunOnce(ctx)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if sent != 0 {
		t.Errorf("second run sent %d reminders, want 0", sent)
	}
}
