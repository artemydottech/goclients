package repository

import (
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(c models.Client) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO clients (company_id, name, phone, email, comment) VALUES (?, ?, ?, ?, ?)",
		c.CompanyID, c.Name, c.Phone, c.Email, c.Comment,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *ClientRepository) GetAllClients() ([]models.Client, error) {
	return r.queryClients(`
        SELECT id, company_id, name, phone, email, comment
        FROM clients ORDER BY id`)
}

func (r *ClientRepository) GetClientsByCompany(companyID int) ([]models.Client, error) {
	return r.queryClients(`
        SELECT id, company_id, name, phone, email, comment
        FROM clients WHERE company_id = ? ORDER BY id`, companyID)
}

func (r *ClientRepository) queryClients(query string, args ...any) ([]models.Client, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		var c models.Client
		err := rows.Scan(&c.ID, &c.CompanyID, &c.Name, &c.Phone, &c.Email, &c.Comment)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}

	return clients, rows.Err()
}

func (r *ClientRepository) GetClientById(id int) (models.Client, error) {
	var c models.Client

	err := r.db.QueryRow(`
        SELECT id, company_id, name, phone, email, comment
        FROM clients WHERE id = ?`, id).
		Scan(&c.ID, &c.CompanyID, &c.Name, &c.Phone, &c.Email, &c.Comment)
	if err != nil {
		return models.Client{}, err
	}

	return c, nil
}

func (r *ClientRepository) GetClientByPhone(companyID int, phone string) (models.Client, error) {
	var c models.Client

	err := r.db.QueryRow(`
        SELECT id, company_id, name, phone, email, comment
        FROM clients WHERE company_id = ? AND phone = ?`, companyID, phone).
		Scan(&c.ID, &c.CompanyID, &c.Name, &c.Phone, &c.Email, &c.Comment)
	if err != nil {
		return models.Client{}, err
	}

	return c, nil
}

func (r *ClientRepository) DeleteClientById(id int) error {
	res, err := r.db.Exec("DELETE FROM clients WHERE id = ?", id)
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
