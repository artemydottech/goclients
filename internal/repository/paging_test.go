package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func ids[T any](items []T, id func(T) int) []int {
	out := make([]int, 0, len(items))
	for _, item := range items {
		out = append(out, id(item))
	}

	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestCompanyRepositoryPaging(t *testing.T) {
	repo := NewCompanyRepository(newTestDB(t))
	ctx := context.Background()

	var created []int
	for i := 1; i <= 5; i++ {
		id, err := repo.Create(ctx, models.Company{Name: fmt.Sprintf("Company %d", i), Timezone: "UTC"})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		created = append(created, int(id))
	}

	tests := []struct {
		name string
		page models.Page
		want []int
	}{
		{"everything", models.Page{}, created},
		{"first two", models.Page{Limit: 2}, created[:2]},
		{"middle", models.Page{Limit: 2, Offset: 2}, created[2:4]},
		{"tail is shorter than the limit", models.Page{Limit: 3, Offset: 4}, created[4:]},
		{"offset only", models.Page{Offset: 3}, created[3:]},
		{"past the end", models.Page{Limit: 2, Offset: 10}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetAllCompanies(ctx, tt.page)
			if err != nil {
				t.Fatalf("GetAllCompanies: %v", err)
			}

			if gotIDs := ids(got, func(c models.Company) int { return c.ID }); !equalInts(gotIDs, tt.want) {
				t.Errorf("ids = %v, want %v", gotIDs, tt.want)
			}
		})
	}
}

func TestEmployeeRepositoryPaging(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	companyID, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: "Romashka", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	repo := NewEmployeeRepository(db)
	var created []int
	for i := 1; i <= 4; i++ {
		id, err := repo.Create(ctx, models.Employee{CompanyID: int(companyID), Name: fmt.Sprintf("Name %d", i), Surname: "Surname"})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		created = append(created, int(id))
	}

	got, err := repo.GetAllEmployees(ctx, models.Page{Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("GetAllEmployees: %v", err)
	}

	if gotIDs := ids(got, func(e models.Employee) int { return e.ID }); !equalInts(gotIDs, created[1:3]) {
		t.Errorf("ids = %v, want %v", gotIDs, created[1:3])
	}
}

func TestServiceRepositoryPagingKeepsTheCompanyFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	companies := NewCompanyRepository(db)
	first, err := companies.Create(ctx, models.Company{Name: "First", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	second, err := companies.Create(ctx, models.Company{Name: "Second", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	repo := NewServiceRepository(db)
	var mine []int
	for i := 1; i <= 4; i++ {
		id, err := repo.Create(ctx, models.Service{CompanyID: int(first), Name: fmt.Sprintf("Mine %d", i), Duration: 30})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		mine = append(mine, int(id))
		if _, err := repo.Create(ctx, models.Service{CompanyID: int(second), Name: fmt.Sprintf("Other %d", i), Duration: 30}); err != nil {
			t.Fatalf("create other: %v", err)
		}
	}

	got, err := repo.GetServicesByCompany(ctx, int(first), models.Page{Limit: 3, Offset: 2})
	if err != nil {
		t.Fatalf("GetServicesByCompany: %v", err)
	}
	if gotIDs := ids(got, func(s models.Service) int { return s.ID }); !equalInts(gotIDs, mine[2:]) {
		t.Errorf("ids = %v, want %v", gotIDs, mine[2:])
	}

	all, err := repo.GetAllServices(ctx, models.Page{Limit: 5})
	if err != nil {
		t.Fatalf("GetAllServices: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("got %d services, want 5", len(all))
	}
}

func TestClientRepositoryPagingKeepsTheCompanyFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	companies := NewCompanyRepository(db)
	first, err := companies.Create(ctx, models.Company{Name: "First", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	second, err := companies.Create(ctx, models.Company{Name: "Second", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	repo := NewClientRepository(db)
	var mine []int
	for i := 1; i <= 4; i++ {
		phone := fmt.Sprintf("7999000000%d", i)
		id, err := repo.Create(ctx, models.Client{CompanyID: int(first), Name: fmt.Sprintf("Mine %d", i), Phone: phone})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		mine = append(mine, int(id))
		if _, err := repo.Create(ctx, models.Client{CompanyID: int(second), Name: fmt.Sprintf("Other %d", i), Phone: phone}); err != nil {
			t.Fatalf("create other: %v", err)
		}
	}

	got, err := repo.GetClientsByCompany(ctx, int(first), models.Page{Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("GetClientsByCompany: %v", err)
	}
	if gotIDs := ids(got, func(c models.Client) int { return c.ID }); !equalInts(gotIDs, mine[1:3]) {
		t.Errorf("ids = %v, want %v", gotIDs, mine[1:3])
	}

	all, err := repo.GetAllClients(ctx, models.Page{Limit: 3})
	if err != nil {
		t.Fatalf("GetAllClients: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("got %d clients, want 3", len(all))
	}
}
