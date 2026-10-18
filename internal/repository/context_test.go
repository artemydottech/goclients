package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func assertCancelled(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestUserRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewUserRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.User{Name: "Anna", Surname: "Ivanova", Username: "anna"})
	assertCancelled(t, err)

	_, err = repo.GetAllUsers(ctx)
	assertCancelled(t, err)

	_, err = repo.GetUserById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteUserById(ctx, 1))
}

func TestCompanyRepositoryHonorsCancelledContext(t *testing.T) {
	repo := NewCompanyRepository(newTestDB(t))
	ctx := cancelledContext()

	_, err := repo.Create(ctx, models.Company{Name: "Romashka"})
	assertCancelled(t, err)

	_, err = repo.GetAllCompanies(ctx)
	assertCancelled(t, err)

	_, err = repo.GetCompanyById(ctx, 1)
	assertCancelled(t, err)

	assertCancelled(t, repo.DeleteCompanyById(ctx, 1))
}
