package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/artemydottech/goclients/internal/models"
)

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func assertCancelled(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestUserRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewUserRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.User{Name: "Anna", Surname: "Ivanova", Username: "anna"})
	assertCancelled(t, err)

	_, err = repo.GetAllUsers(ctx)
	assertCancelled(t, err)

	_, err = repo.GetUserById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteUserById(ctx, 1))
}

func TestCompanyRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewCompanyRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.Company{Name: "Romashka"})
	assertCancelled(t, err)

	_, err = repo.GetAllCompanies(ctx)
	assertCancelled(t, err)

	_, err = repo.GetCompanyById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteCompanyById(ctx, 1))
}

func TestEmployeeRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewEmployeeRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.Employee{CompanyID: 1, Name: "Anna", Surname: "Ivanova"})
	assertCancelled(t, err)

	_, err = repo.GetAllEmployees(ctx)
	assertCancelled(t, err)

	_, err = repo.GetEmployeeById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteEmployeeById(ctx, 1))
}

func TestServiceRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewServiceRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.Service{CompanyID: 1, Name: "Haircut", Duration: 30})
	assertCancelled(t, err)

	_, err = repo.GetAllServices(ctx)
	assertCancelled(t, err)

	_, err = repo.GetServicesByCompany(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetServiceById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteServiceById(ctx, 1))
}

func TestClientRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewClientRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.Client{CompanyID: 1, Name: "Ivan", Phone: "+79990000000"})
	assertCancelled(t, err)

	_, err = repo.GetAllClients(ctx)
	assertCancelled(t, err)

	_, err = repo.GetClientsByCompany(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetClientById(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetClientByPhone(ctx, 1, "+79990000000")
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteClientById(ctx, 1))
}

func TestAssignmentRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewAssignmentRepository(newTestDB(t))
	ctx := cancelledContext()

	assertCancelled(t, repo.SetEmployeeServices(ctx, 1, []int{1, 2}))

	_, err := repo.GetServicesByEmployee(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetEmployeesByService(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.EmployeePerformsService(ctx, 1, 1)
	assertCancelled(t, err)
}

func TestScheduleRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewScheduleRepository(newTestDB(t))
	ctx := cancelledContext()

	days := []models.WorkingDay{{Weekday: 1, StartsAt: "09:00", EndsAt: "18:00"}}
	assertCancelled(t, repo.SetEmployeeSchedule(ctx, 1, days))

	_, err := repo.GetEmployeeSchedule(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetWorkingDay(ctx, 1, 1)
	assertCancelled(t, err)
}

func TestTimeOffRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewTimeOffRepository(newTestDB(t))
	ctx := cancelledContext()
	from := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	_, err := repo.Create(ctx, models.TimeOff{EmployeeID: 1, StartsAt: from, EndsAt: from.AddDate(0, 0, 3)})
	assertCancelled(t, err)

	_, err = repo.GetTimeOffByEmployee(ctx, 1)
	assertCancelled(t, err)

	_, err = repo.GetTimeOffInRange(ctx, 1, from, from.AddDate(0, 0, 7))
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteTimeOffById(ctx, 1))
}
