package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

type fakeUserServ struct {
	registerID  int64
	registerErr error
	users       []models.User
	listErr     error
	user        models.User
	getErr      error
	deleteErr   error
	registered  models.User
}

func (f *fakeUserServ) RegisterUser(ctx context.Context, u models.User) (int64, error) {
	f.registered = u
	return f.registerID, f.registerErr
}

func (f *fakeUserServ) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return f.users, f.listErr
}

func (f *fakeUserServ) GetUserById(ctx context.Context, id int) (models.User, error) {
	return f.user, f.getErr
}

func (f *fakeUserServ) DeleteUserById(ctx context.Context, id int) error {
	return f.deleteErr
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serv       *fakeUserServ
		wantStatus int
	}{
		{"created", `{"name":"Anna","username":"anna"}`, &fakeUserServ{registerID: 5}, http.StatusCreated},
		{"malformed body", `{oops`, &fakeUserServ{}, http.StatusBadRequest},
		{"validation error", `{"name":""}`, &fakeUserServ{registerErr: models.Invalid("name is required")}, http.StatusBadRequest},
		{"storage failure", `{"name":"Anna"}`, &fakeUserServ{registerErr: errors.New("disk full")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.serv)

			rec := serve(t, "POST /users", h.CreateUser, http.MethodPost, "/users", tt.body)

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestCreateUserReturnsID(t *testing.T) {
	serv := &fakeUserServ{registerID: 42}
	h := NewUserHandler(serv)

	rec := serve(t, "POST /users", h.CreateUser, http.MethodPost, "/users", `{"name":"Anna","username":"anna"}`)

	assertStatus(t, rec, http.StatusCreated)
	var body map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != 42 {
		t.Errorf("id = %d, want 42", body["id"])
	}
	if serv.registered.Username != "anna" {
		t.Errorf("service got %+v", serv.registered)
	}
}

func TestGetAllUsers(t *testing.T) {
	h := NewUserHandler(&fakeUserServ{users: []models.User{{ID: 1, Name: "Anna"}, {ID: 2, Name: "Ivan"}}})

	rec := serve(t, "GET /users", h.GetAllUsers, http.MethodGet, "/users", "")

	assertStatus(t, rec, http.StatusOK)
	var users []models.User
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("got %d users, want 2", len(users))
	}
}

func TestGetAllUsersStorageFailure(t *testing.T) {
	h := NewUserHandler(&fakeUserServ{listErr: errors.New("boom")})

	rec := serve(t, "GET /users", h.GetAllUsers, http.MethodGet, "/users", "")

	assertStatus(t, rec, http.StatusInternalServerError)
}

func TestGetUserById(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeUserServ
		wantStatus int
	}{
		{"found", "/users/3", &fakeUserServ{user: models.User{ID: 3}}, http.StatusOK},
		{"not found", "/users/3", &fakeUserServ{getErr: models.ErrNotFound}, http.StatusNotFound},
		{"wrapped not found", "/users/3", &fakeUserServ{getErr: errors.Join(errors.New("ctx"), models.ErrNotFound)}, http.StatusNotFound},
		{"bad id", "/users/abc", &fakeUserServ{}, http.StatusBadRequest},
		{"storage failure", "/users/3", &fakeUserServ{getErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.serv)

			rec := serve(t, "GET /users/{id}", h.GetUserById, http.MethodGet, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serv       *fakeUserServ
		wantStatus int
	}{
		{"deleted", "/users/3", &fakeUserServ{}, http.StatusNoContent},
		{"not found", "/users/3", &fakeUserServ{deleteErr: models.ErrNotFound}, http.StatusNotFound},
		{"bad id", "/users/x", &fakeUserServ{}, http.StatusBadRequest},
		{"storage failure", "/users/3", &fakeUserServ{deleteErr: errors.New("boom")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.serv)

			rec := serve(t, "DELETE /users/{id}", h.DeleteUser, http.MethodDelete, tt.target, "")

			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
