package repository

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}

func TestUserRepositoryRoundTrip(t *testing.T) {
	repo := NewUserRepository(newTestDB(t))

	want := models.User{Name: "Артемий", Surname: "Зверев", Username: "artemy", Avatar: "a.png"}

	id, err := repo.Create(want)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, err := repo.GetUserById(int(id))
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	want.ID = int(id)
	if user != want {
		t.Errorf("got %+v, want %+v", user, want)
	}

	all, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("got %d users, want 1", len(all))
	}

	if err := repo.DeleteUserById(int(id)); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.DeleteUserById(int(id)); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("second delete: got %v, want sql.ErrNoRows", err)
	}
}

func TestUserRepositoryMissingUser(t *testing.T) {
	repo := NewUserRepository(newTestDB(t))

	if _, err := repo.GetUserById(404); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got %v, want sql.ErrNoRows", err)
	}
}

func TestCompanyRepositoryKeepsSocials(t *testing.T) {
	repo := NewCompanyRepository(newTestDB(t))

	id, err := repo.Create(models.Company{
		Name:    "Rusholding",
		Address: "Екатеринбург",
		Socials: models.Socials{models.SocialTelegram: "@rusholding"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	company, err := repo.GetCompanyById(int(id))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got := company.Socials[models.SocialTelegram]; got != "@rusholding" {
		t.Errorf("telegram = %q", got)
	}

	all, err := repo.GetAllCompanies()
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(all) != 1 || all[0].Socials[models.SocialTelegram] != "@rusholding" {
		t.Errorf("got %+v", all)
	}
}

func TestDeletingACompanyRemovesItsEmployees(t *testing.T) {
	db := newTestDB(t)
	companies := NewCompanyRepository(db)
	employees := NewEmployeeRepository(db)

	companyID, err := companies.Create(models.Company{Name: "Rusholding"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	if _, err := employees.Create(models.Employee{
		CompanyID: int(companyID),
		Name:      "Артемий",
		Surname:   "Зверев",
	}); err != nil {
		t.Fatalf("create employee: %v", err)
	}

	if err := companies.DeleteCompanyById(int(companyID)); err != nil {
		t.Fatalf("delete company: %v", err)
	}

	left, err := employees.GetAllEmployees()
	if err != nil {
		t.Fatalf("get employees: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("company gone but %d employees left behind", len(left))
	}
}

func TestMigrateAddsTheNewUserColumnsToAnOldDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatalf("old schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (name) VALUES ('Артемий')`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	users, err := NewUserRepository(db).GetAllUsers()
	if err != nil {
		t.Fatalf("read after migrate: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Артемий" || users[0].Surname != "" {
		t.Errorf("got %+v", users)
	}

	if err := Migrate(db); err != nil {
		t.Errorf("second migrate: %v", err)
	}
}

func TestServiceRepositoryRoundTrip(t *testing.T) {
	db := newTestDB(t)
	companies := NewCompanyRepository(db)
	services := NewServiceRepository(db)

	companyID, err := companies.Create(models.Company{Name: "Ромашка"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	want := models.Service{
		CompanyID:   int(companyID),
		Name:        "Стрижка",
		Description: "Мужская",
		Duration:    45,
		Price:       1500,
	}

	id, err := services.Create(want)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	got, err := services.GetServiceById(int(id))
	if err != nil {
		t.Fatalf("get service: %v", err)
	}

	want.ID = int(id)
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	byCompany, err := services.GetServicesByCompany(int(companyID))
	if err != nil {
		t.Fatalf("get by company: %v", err)
	}
	if len(byCompany) != 1 {
		t.Fatalf("услуг компании: %d, ожидалась 1", len(byCompany))
	}

	other, err := services.GetServicesByCompany(int(companyID) + 1000)
	if err != nil {
		t.Fatalf("get by unknown company: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("чужая компания вернула %d услуг", len(other))
	}
}

func TestServicesGoAwayWithTheirCompany(t *testing.T) {
	db := newTestDB(t)
	companies := NewCompanyRepository(db)
	services := NewServiceRepository(db)

	companyID, err := companies.Create(models.Company{Name: "Ромашка"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	serviceID, err := services.Create(models.Service{
		CompanyID: int(companyID), Name: "Стрижка", Duration: 45,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	if err := companies.DeleteCompanyById(int(companyID)); err != nil {
		t.Fatalf("delete company: %v", err)
	}

	if _, err := services.GetServiceById(int(serviceID)); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("услуга пережила удаление компании: %v", err)
	}
}

func TestAssignmentRepositoryReplacesTheWholeSet(t *testing.T) {
	db := newTestDB(t)
	companies := NewCompanyRepository(db)
	employees := NewEmployeeRepository(db)
	services := NewServiceRepository(db)
	assignments := NewAssignmentRepository(db)

	companyID, err := companies.Create(models.Company{Name: "Ромашка"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	employeeID, err := employees.Create(models.Employee{
		CompanyID: int(companyID), Name: "Анна", Surname: "Иванова",
	})
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	first, err := services.Create(models.Service{
		CompanyID: int(companyID), Name: "Стрижка", Duration: 45,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	second, err := services.Create(models.Service{
		CompanyID: int(companyID), Name: "Окрашивание", Duration: 120,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	if err := assignments.SetEmployeeServices(int(employeeID), []int{int(first), int(second)}); err != nil {
		t.Fatalf("set services: %v", err)
	}

	got, err := assignments.GetServicesByEmployee(int(employeeID))
	if err != nil {
		t.Fatalf("get services: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("услуг сотрудника: %d, ожидалось 2", len(got))
	}

	if err := assignments.SetEmployeeServices(int(employeeID), []int{int(second)}); err != nil {
		t.Fatalf("replace services: %v", err)
	}

	got, err = assignments.GetServicesByEmployee(int(employeeID))
	if err != nil {
		t.Fatalf("get services: %v", err)
	}
	if len(got) != 1 || got[0].ID != int(second) {
		t.Fatalf("после замены получено %+v", got)
	}

	masters, err := assignments.GetEmployeesByService(int(second))
	if err != nil {
		t.Fatalf("get employees: %v", err)
	}
	if len(masters) != 1 || masters[0].ID != int(employeeID) {
		t.Fatalf("мастера услуги: %+v", masters)
	}

	performs, err := assignments.EmployeePerformsService(int(employeeID), int(first))
	if err != nil {
		t.Fatalf("performs: %v", err)
	}
	if performs {
		t.Error("снятая услуга всё ещё числится за сотрудником")
	}
}

func TestAssignmentsGoAwayWithTheService(t *testing.T) {
	db := newTestDB(t)
	companies := NewCompanyRepository(db)
	employees := NewEmployeeRepository(db)
	services := NewServiceRepository(db)
	assignments := NewAssignmentRepository(db)

	companyID, err := companies.Create(models.Company{Name: "Ромашка"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	employeeID, err := employees.Create(models.Employee{
		CompanyID: int(companyID), Name: "Анна", Surname: "Иванова",
	})
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	serviceID, err := services.Create(models.Service{
		CompanyID: int(companyID), Name: "Стрижка", Duration: 45,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	if err := assignments.SetEmployeeServices(int(employeeID), []int{int(serviceID)}); err != nil {
		t.Fatalf("set services: %v", err)
	}

	if err := services.DeleteServiceById(int(serviceID)); err != nil {
		t.Fatalf("delete service: %v", err)
	}

	got, err := assignments.GetServicesByEmployee(int(employeeID))
	if err != nil {
		t.Fatalf("get services: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("связка пережила удаление услуги: %+v", got)
	}
}

// bookingFixture заводит компанию, мастера, услугу и клиента — минимум, без
// которого запись не создать.
func bookingFixture(t *testing.T, db *sql.DB) (companyID, employeeID, serviceID, clientID int) {
	t.Helper()

	company, err := NewCompanyRepository(db).Create(models.Company{Name: "Ромашка"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	employee, err := NewEmployeeRepository(db).Create(models.Employee{
		CompanyID: int(company), Name: "Анна", Surname: "Иванова",
	})
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	service, err := NewServiceRepository(db).Create(models.Service{
		CompanyID: int(company), Name: "Стрижка", Duration: 60,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	client, err := NewClientRepository(db).Create(models.Client{
		CompanyID: int(company), Name: "Пётр", Phone: "79991234567",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	return int(company), int(employee), int(service), int(client)
}

func TestAppointmentRepositoryRoundTrip(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	want := models.Appointment{
		CompanyID:  companyID,
		ClientID:   clientID,
		EmployeeID: employeeID,
		ServiceID:  serviceID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.Add(time.Hour),
		Status:     models.AppointmentPending,
		Comment:    "первый визит",
	}

	id, err := repo.Create(want)
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	got, err := repo.GetAppointmentById(int(id))
	if err != nil {
		t.Fatalf("get appointment: %v", err)
	}

	want.ID = int(id)
	if !got.StartsAt.Equal(want.StartsAt) || !got.EndsAt.Equal(want.EndsAt) {
		t.Errorf("время не сходится: %v–%v против %v–%v", got.StartsAt, got.EndsAt, want.StartsAt, want.EndsAt)
	}
	if got.Status != want.Status || got.Comment != want.Comment {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestHasOverlapTreatsTouchingIntervalsAsFree(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	if _, err := repo.Create(models.Appointment{
		CompanyID: companyID, ClientID: clientID, EmployeeID: employeeID, ServiceID: serviceID,
		StartsAt: startsAt, EndsAt: startsAt.Add(time.Hour), Status: models.AppointmentConfirmed,
	}); err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	cases := map[string]struct {
		from, to time.Time
		want     bool
	}{
		"впритык после":  {startsAt.Add(time.Hour), startsAt.Add(2 * time.Hour), false},
		"впритык до":     {startsAt.Add(-time.Hour), startsAt, false},
		"внахлёст":       {startsAt.Add(30 * time.Minute), startsAt.Add(90 * time.Minute), true},
		"целиком внутри": {startsAt.Add(10 * time.Minute), startsAt.Add(20 * time.Minute), true},
		"накрывает":      {startsAt.Add(-time.Hour), startsAt.Add(2 * time.Hour), true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			busy, err := repo.HasOverlap(employeeID, tc.from, tc.to)
			if err != nil {
				t.Fatalf("overlap: %v", err)
			}
			if busy != tc.want {
				t.Errorf("HasOverlap = %v, ожидалось %v", busy, tc.want)
			}
		})
	}
}

func TestCancelledAppointmentFreesTheSlot(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	id, err := repo.Create(models.Appointment{
		CompanyID: companyID, ClientID: clientID, EmployeeID: employeeID, ServiceID: serviceID,
		StartsAt: startsAt, EndsAt: startsAt.Add(time.Hour), Status: models.AppointmentConfirmed,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	busy, err := repo.HasOverlap(employeeID, startsAt, startsAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("overlap: %v", err)
	}
	if !busy {
		t.Fatal("подтверждённая запись должна занимать слот")
	}

	if err := repo.UpdateStatus(int(id), models.AppointmentCancelled); err != nil {
		t.Fatalf("update status: %v", err)
	}

	busy, err = repo.HasOverlap(employeeID, startsAt, startsAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("overlap: %v", err)
	}
	if busy {
		t.Error("после отмены слот должен освободиться")
	}
}

func TestAppointmentsGoAwayWithTheClient(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	id, err := repo.Create(models.Appointment{
		CompanyID: companyID, ClientID: clientID, EmployeeID: employeeID, ServiceID: serviceID,
		StartsAt: startsAt, EndsAt: startsAt.Add(time.Hour), Status: models.AppointmentPending,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	if err := NewClientRepository(db).DeleteClientById(clientID); err != nil {
		t.Fatalf("delete client: %v", err)
	}

	if _, err := repo.GetAppointmentById(int(id)); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("запись пережила удаление клиента: %v", err)
	}
}

func TestCreateIfFreeLetsOnlyOneConcurrentBookingThrough(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	booking := models.Appointment{
		CompanyID: companyID, ClientID: clientID, EmployeeID: employeeID, ServiceID: serviceID,
		StartsAt: startsAt, EndsAt: startsAt.Add(time.Hour), Status: models.AppointmentPending,
	}

	const attempts = 20

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		created int
	)

	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, busy, err := repo.CreateIfFree(booking)
			if err != nil {
				t.Errorf("create if free: %v", err)
				return
			}
			if !busy {
				mu.Lock()
				created++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if created != 1 {
		t.Fatalf("на один слот создано %d записей, ожидалась 1", created)
	}

	all, err := repo.GetAllAppointments()
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("в базе %d записей, ожидалась 1", len(all))
	}
}

func TestCompanyRepositoryKeepsTimezone(t *testing.T) {
	repo := NewCompanyRepository(newTestDB(t))

	id, err := repo.Create(models.Company{Name: "Ромашка", Timezone: "Asia/Yekaterinburg"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	company, err := repo.GetCompanyById(int(id))
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if company.Timezone != "Asia/Yekaterinburg" {
		t.Errorf("timezone %q, ожидался Asia/Yekaterinburg", company.Timezone)
	}
}

func TestMigrateAddsTimezoneToAnOldCompaniesTable(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "old.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`CREATE TABLE companies (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        address TEXT, geolocation TEXT, schedule TEXT, site TEXT, socials TEXT, logo TEXT
    )`); err != nil {
		t.Fatalf("old schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO companies (name, address, geolocation, schedule, site, socials, logo)
        VALUES ('Старая', '', '', '', '', 'null', '')`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	company, err := NewCompanyRepository(db).GetCompanyById(1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if company.Timezone != "UTC" {
		t.Errorf("старой компании достался пояс %q, ожидался UTC", company.Timezone)
	}
}

func TestMoveIfFreeIgnoresTheAppointmentBeingMoved(t *testing.T) {
	db := newTestDB(t)
	companyID, employeeID, serviceID, clientID := bookingFixture(t, db)
	repo := NewAppointmentRepository(db)

	startsAt := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	base := models.Appointment{
		CompanyID: companyID, ClientID: clientID, EmployeeID: employeeID, ServiceID: serviceID,
		Status: models.AppointmentConfirmed,
	}

	own := base
	own.StartsAt, own.EndsAt = startsAt, startsAt.Add(time.Hour)
	ownID, err := repo.Create(own)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	other := base
	other.StartsAt, other.EndsAt = startsAt.Add(2*time.Hour), startsAt.Add(3*time.Hour)
	if _, err := repo.Create(other); err != nil {
		t.Fatalf("create: %v", err)
	}

	shifted := own
	shifted.StartsAt, shifted.EndsAt = startsAt.Add(30*time.Minute), startsAt.Add(90*time.Minute)
	busy, err := repo.MoveIfFree(int(ownID), shifted)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if busy {
		t.Fatal("сдвиг внутри собственного интервала посчитан конфликтом")
	}

	got, err := repo.GetAppointmentById(int(ownID))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.StartsAt.Equal(shifted.StartsAt) {
		t.Errorf("запись осталась на %v", got.StartsAt)
	}

	clash := own
	clash.StartsAt, clash.EndsAt = startsAt.Add(150*time.Minute), startsAt.Add(210*time.Minute)
	busy, err = repo.MoveIfFree(int(ownID), clash)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if !busy {
		t.Error("перенос на чужую запись должен отклоняться")
	}

	free := own
	free.StartsAt, free.EndsAt = startsAt.Add(5*time.Hour), startsAt.Add(6*time.Hour)
	if _, err := repo.MoveIfFree(9999, free); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("перенос несуществующей записи: %v", err)
	}
}

func TestTimeOffRepositoryRangeLookup(t *testing.T) {
	db := newTestDB(t)
	_, employeeID, _, _ := bookingFixture(t, db)
	repo := NewTimeOffRepository(db)

	startsAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	id, err := repo.Create(models.TimeOff{
		EmployeeID: employeeID,
		StartsAt:   startsAt,
		EndsAt:     startsAt.AddDate(0, 0, 7),
		Reason:     "отпуск",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	all, err := repo.GetTimeOffByEmployee(employeeID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(all) != 1 || all[0].Reason != "отпуск" || !all[0].StartsAt.Equal(startsAt) {
		t.Fatalf("получено %+v", all)
	}

	inside, err := repo.GetTimeOffInRange(employeeID, startsAt.AddDate(0, 0, 3), startsAt.AddDate(0, 0, 4))
	if err != nil {
		t.Fatalf("range: %v", err)
	}
	if len(inside) != 1 {
		t.Errorf("день внутри отпуска не нашёл период: %+v", inside)
	}

	after, err := repo.GetTimeOffInRange(employeeID, startsAt.AddDate(0, 0, 7), startsAt.AddDate(0, 0, 8))
	if err != nil {
		t.Fatalf("range: %v", err)
	}
	if len(after) != 0 {
		t.Errorf("день сразу после отпуска попал в период: %+v", after)
	}

	if err := repo.DeleteTimeOffById(int(id)); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.DeleteTimeOffById(int(id)); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("повторное удаление: %v", err)
	}
}

func TestScheduleRepositoryKeepsTheBreak(t *testing.T) {
	db := newTestDB(t)
	_, employeeID, _, _ := bookingFixture(t, db)
	repo := NewScheduleRepository(db)

	want := models.WorkingDay{
		EmployeeID: employeeID, Weekday: 2, StartsAt: "10:00", EndsAt: "20:00",
		BreakStartsAt: "13:00", BreakEndsAt: "14:00",
	}

	if err := repo.SetEmployeeSchedule(employeeID, []models.WorkingDay{want}); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := repo.GetWorkingDay(employeeID, 2)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
