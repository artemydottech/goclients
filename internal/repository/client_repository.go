package repository

import (
	"context"
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, c models.Client) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO clients (company_id, name, phone, email, comment) VALUES (?, ?, ?, ?, ?)",
		c.CompanyID, c.Name, c.Phone, c.Email, c.Comment,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *ClientRepository) GetAllClients(ctx context.Context) ([]models.Client, error) {
	return r.queryClients(ctx, `
        SELECT id, company_id, name, phone, email, comment
        FROM clients ORDER BY id`)
}

func (r *ClientRepository) GetClientsByCompany(ctx context.Context, companyID int) ([]models.Client, error) {
	return r.queryClients(ctx, `
        SELECT id, company_id, name, phone, email, comment
        FROM clients WHERE company_id = ? ORDER BY id`, companyID)
}

func (r *ClientRepository) queryClients(ctx context.Context, query string, args ...any) ([]models.Client, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
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
		return models.Client{}, mapNoRows(err)
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
		return models.Client{}, mapNoRows(err)
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
		return errNoRows
	}

	return nil
}
