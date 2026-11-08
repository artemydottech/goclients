package repository

import (
	"context"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestClientSearchIsolationAndExactPhone(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewClientRepository(db)
	for _, name := range []string{"A", "B"} {
		id, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: name})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Create(ctx, models.Client{CompanyID: int(id), Name: "Анна", Phone: "79990000000"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.SearchClients(ctx, models.ClientSearch{CompanyID: 1, Phone: "79990000000"})
	if err != nil || len(got) != 1 || got[0].CompanyID != 1 {
		t.Fatalf("search: %+v %v", got, err)
	}
	got, err = repo.SearchClients(ctx, models.ClientSearch{CompanyID: 1, Phone: "79990000001"})
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("missing: %+v %v", got, err)
	}
}
func TestClientSearchUnicodeLiteralAndPaging(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewClientRepository(db)
	id, err := NewCompanyRepository(db).Create(ctx, models.Company{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"Анна", "АННА", "100%_special", "Other"} {
		phone := []string{"79990000000", "79990000001", "79990000002", "79990000003"}[i]
		if _, err := repo.Create(ctx, models.Client{CompanyID: int(id), Name: name, Phone: phone, Email: "USER@example.com"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		q     string
		p     models.Page
		count int
	}{{"анна", models.Page{}, 2}, {"АННА", models.Page{Limit: 1, Offset: 1}, 1}, {"%_", models.Page{}, 1}, {"' OR 1=1", models.Page{}, 0}, {"example", models.Page{}, 4}, {"00003", models.Page{}, 1}, {"+7 (999) 000-00-03", models.Page{}, 1}} {
		got, err := repo.SearchClients(ctx, models.ClientSearch{CompanyID: int(id), Query: tt.q, Page: tt.p})
		if err != nil || len(got) != tt.count {
			t.Fatalf("%q: %+v %v", tt.q, got, err)
		}
	}
}
