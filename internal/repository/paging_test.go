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
