package repository

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
)

func TestEmployeeCompanyIsolation(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewEmployeeRepository(db)
	a, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{int(a), int(b), int(a)} {
		if _, err := repo.Create(ctx, models.Employee{CompanyID: id, Name: "Anna", Surname: "Ivanova"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.GetEmployeesByCompany(ctx, int(a), models.Page{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
	for _, e := range got {
		if e.CompanyID != int(a) {
			t.Fatalf("foreign employee: %+v", e)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.GetEmployeesByCompany(cancelled, int(a), models.Page{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
