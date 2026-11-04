package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"testing"
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
