package repository

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

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

	id, err := repo.Create("Артемий")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, err := repo.GetUserById(int(id))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if user.Name != "Артемий" {
		t.Errorf("name = %q", user.Name)
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
