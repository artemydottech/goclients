package repository

import (
	"database/sql"
	"log"

	"github.com/artemydottech/goclients/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(name string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := r.db.Query("SELECT id, name FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.ID, &user.Name)
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

func (r *UserRepository) GetUserById(id int) (models.User, error) {
	var user models.User

	err := r.db.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) DeleteUserById(id int) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
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

func (r *UserRepository) TestRows() {
	rows, _ := r.db.Query("SELECT id, name FROM users")

	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		log.Printf("Строка #%d: ID=%d Name=%s", count, id, name)
		count++
	}
	log.Printf("Всего строк: %d", count)
}
