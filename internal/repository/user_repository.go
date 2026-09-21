package repository

import (
	"context"
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u models.User) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users (name, surname, username, avatar) VALUES (?, ?, ?, ?)",
		u.Name, u.Surname, u.Username, u.Avatar,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, surname, username, avatar FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []models.User{}

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.Username, &user.Avatar)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, id int) (models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(ctx, "SELECT id, name, surname, username, avatar FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Surname, &user.Username, &user.Avatar)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) DeleteUserById(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
