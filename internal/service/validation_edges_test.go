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

func TestValidationHarnessImports(t *testing.T) {
	_ = context.Background()
	_ = math.NaN()
}
