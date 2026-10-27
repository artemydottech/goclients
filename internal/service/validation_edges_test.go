package service

import (
	"context"
	"errors"
	"github.com/artemydottech/goclients/internal/models"
	"math"
	"strings"
	"testing"
)

func assertValidation(t *testing.T, err error) {
	t.Helper()
	var target models.ValidationError
	if !errors.As(err, &target) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestClientRejectsMalformedMailbox(t *testing.T) {
	for _, email := range []string{"@", "anna@", "@example.com", "Anna <anna@example.com>", "a@example.com,b@example.com", strings.Repeat("a", 255) + "@example.com"} {
		c := validClient()
		c.Email = email
		_, err := validateClient(c)
		assertValidation(t, err)
	}
	c := validClient()
	c.Email = "  anna@example.com  "
	got, err := validateClient(c)
	if err != nil || got.Email != "anna@example.com" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestServiceRejectsNonfinitePrices(t *testing.T) {
	for _, price := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		item := validService()
		item.Price = price
		assertValidation(t, validateService(item))
	}
}

func TestUserRejectsBlankNameAndOversizedAvatar(t *testing.T) {
	for _, user := range []models.User{{Name: " \t\n"}, {Name: "Anna", Avatar: strings.Repeat("a", 501)}} {
		_, err := NewUserService(&stubUserRepo{}).RegisterUser(context.Background(), user)
		assertValidation(t, err)
	}
}

func TestEmployeeRequiresACompanyBeforeInsertion(t *testing.T) {
	for _, id := range []int{0, -1} {
		repo := &stubEmployeeRepo{}
		_, err := NewEmployeeService(repo).CreateEmployee(context.Background(), models.Employee{CompanyID: id, Name: "Anna", Surname: "Ivanova"})
		assertValidation(t, err)
		if repo.called {
			t.Fatal("invalid company reached repository")
		}
	}
}

func TestCompanyRejectsUnsafeLinksAndOversizedMetadata(t *testing.T) {
	for _, company := range []models.Company{
		{Name: "Salon", Site: "javascript:alert(1)"},
		{Name: "Salon", Logo: "data:image/svg+xml,test"},
		{Name: "Salon", Site: "https://"},
		{Name: "Salon", Geolocation: strings.Repeat("a", 501)},
		{Name: "Salon", Schedule: strings.Repeat("a", 1001)},
	} {
		_, err := validateCompany(company)
		assertValidation(t, err)
	}
	if _, err := validateCompany(models.Company{Name: "Salon", Site: "https://example.com", Logo: "https://example.com/logo.png"}); err != nil {
		t.Fatal(err)
	}
}
