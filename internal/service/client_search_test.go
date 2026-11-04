package service

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"strings"
	"testing"
)

type searchStub struct {
	called bool
	filter models.ClientSearch
}

func (s *searchStub) SearchClients(ctx context.Context, f models.ClientSearch) ([]models.Client, error) {
	s.called = true
	s.filter = f
	return []models.Client{}, ctx.Err()
}
func TestClientSearchValidationAndNormalization(t *testing.T) {
	for _, f := range []models.ClientSearch{{}, {CompanyID: 1, Query: strings.Repeat("я", 201)}, {CompanyID: 1, Phone: "abc"}, {CompanyID: 1, Phone: "123"}, {CompanyID: 1, Page: models.Page{Limit: 201}}} {
		repo := &searchStub{}
		_, err := NewClientSearchService(repo).SearchClients(context.Background(), f)
		if err == nil || repo.called {
			t.Fatalf("invalid search: %+v", f)
		}
	}
	repo := &searchStub{}
	_, err := NewClientSearchService(repo).SearchClients(context.Background(), models.ClientSearch{CompanyID: 1, Query: " Анна ", Phone: "+7 (999) 000-00-00"})
	if err != nil || repo.filter.Query != "Анна" || repo.filter.Phone != "79990000000" {
		t.Fatalf("normalized: %+v %v", repo.filter, err)
	}
}
