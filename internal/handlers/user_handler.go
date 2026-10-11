package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

type UserServ interface {
	RegisterUser(ctx context.Context, u models.User) (int64, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, id int) (models.User, error)
	DeleteUserById(ctx context.Context, id int) error
}

type UserHandler struct {
	service UserServ
}

func NewUserHandler(s UserServ) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input models.User
	ctx := r.Context()

	if !decodeJSON(w, r, &input) {
		return
	}

	id, err := h.service.RegisterUser(ctx, input)
	var validationErr models.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save user")
		return
	}

	writeStatusJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.service.GetAllUsers(ctx)
	if err != nil {
		log.Printf("GetAllUsers: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load users")
		return
	}

	writeJSON(w, users)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	user, err := h.service.GetUserById(ctx, id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "request failed")
		return
	}

	writeJSON(w, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = h.service.DeleteUserById(ctx, id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
