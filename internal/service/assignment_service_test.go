package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type stubAssignmentRepo struct {
	saved []int
	calls int
}

func (r *stubAssignmentRepo) SetEmployeeServices(_ int, serviceIDs []int) error {
	r.calls++
	r.saved = serviceIDs
	return nil
}

func (r *stubAssignmentRepo) GetServicesByEmployee(int) ([]models.Service, error) {
	return nil, nil
}

func (r *stubAssignmentRepo) GetEmployeesByService(int) ([]models.Employee, error) {
	return nil, nil
}

func (r *stubAssignmentRepo) EmployeePerformsService(int, int) (bool, error) { return false, nil }

type stubEmployeeLookup struct {
	employee models.Employee
	err      error
}

func (l stubEmployeeLookup) GetEmployeeById(int) (models.Employee, error) {
	return l.employee, l.err
}

type stubServiceLookup struct {
	byID map[int]models.Service
}

func (l stubServiceLookup) GetServiceById(id int) (models.Service, error) {
	item, ok := l.byID[id]
	if !ok {
		return models.Service{}, sql.ErrNoRows
	}
	return item, nil
}

func newAssignmentService(repo *stubAssignmentRepo) *AssignmentService {
	return NewAssignmentService(
		repo,
		stubEmployeeLookup{employee: models.Employee{ID: 1, CompanyID: 10}},
		stubServiceLookup{byID: map[int]models.Service{
			100: {ID: 100, CompanyID: 10},
			101: {ID: 101, CompanyID: 10},
			200: {ID: 200, CompanyID: 99},
		}},
	)
}

func TestSetEmployeeServicesSavesTheSet(t *testing.T) {
	repo := &stubAssignmentRepo{}

	if err := newAssignmentService(repo).SetEmployeeServices(1, []int{100, 101}); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(repo.saved) != 2 {
		t.Fatalf("сохранено %d услуг, ожидалось 2", len(repo.saved))
	}
}

func TestSetEmployeeServicesDropsDuplicates(t *testing.T) {
	repo := &stubAssignmentRepo{}

	if err := newAssignmentService(repo).SetEmployeeServices(1, []int{100, 100, 101}); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(repo.saved) != 2 {
		t.Errorf("сохранено %d услуг, дубль должен был отсеяться", len(repo.saved))
	}
}

func TestSetEmployeeServicesRejectsForeignCompany(t *testing.T) {
	repo := &stubAssignmentRepo{}

	err := newAssignmentService(repo).SetEmployeeServices(1, []int{100, 200})

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.calls != 0 {
		t.Error("набор не должен сохраняться, если в нём чужая услуга")
	}
}

func TestSetEmployeeServicesRejectsUnknownService(t *testing.T) {
	repo := &stubAssignmentRepo{}

	err := newAssignmentService(repo).SetEmployeeServices(1, []int{404})

	var validationErr models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ожидалась ValidationError, получено %v", err)
	}

	if repo.calls != 0 {
		t.Error("набор не должен сохраняться при неизвестной услуге")
	}
}

func TestSetEmployeeServicesReportsMissingEmployee(t *testing.T) {
	repo := &stubAssignmentRepo{}

	svc := NewAssignmentService(
		repo,
		stubEmployeeLookup{err: sql.ErrNoRows},
		stubServiceLookup{byID: map[int]models.Service{}},
	)

	if err := svc.SetEmployeeServices(1, []int{100}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ожидалась sql.ErrNoRows, получено %v", err)
	}

	if repo.calls != 0 {
		t.Error("репозиторий не должен вызываться для несуществующего сотрудника")
	}
}

func TestSetEmployeeServicesAcceptsEmptySet(t *testing.T) {
	repo := &stubAssignmentRepo{}

	if err := newAssignmentService(repo).SetEmployeeServices(1, nil); err != nil {
		t.Fatalf("пустой набор снимает все услуги, получено %v", err)
	}

	if repo.calls != 1 {
		t.Error("пустой набор всё равно должен дойти до репозитория")
	}
}
