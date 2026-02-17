package repository

import "database/sql"

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