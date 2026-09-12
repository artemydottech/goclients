package repository

import (
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type ServiceRepository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(s models.Service) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO services (company_id, name, description, duration, price) VALUES (?, ?, ?, ?, ?)",
		s.CompanyID, s.Name, s.Description, s.Duration, s.Price,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *ServiceRepository) GetAllServices() ([]models.Service, error) {
	rows, err := r.db.Query(`
        SELECT id, company_id, name, description, duration, price
        FROM services`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []models.Service{}
	for rows.Next() {
		var s models.Service
		err := rows.Scan(&s.ID, &s.CompanyID, &s.Name, &s.Description, &s.Duration, &s.Price)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *ServiceRepository) GetServicesByCompany(companyID int) ([]models.Service, error) {
	rows, err := r.db.Query(`
        SELECT id, company_id, name, description, duration, price
        FROM services WHERE company_id = ?`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []models.Service{}
	for rows.Next() {
		var s models.Service
		err := rows.Scan(&s.ID, &s.CompanyID, &s.Name, &s.Description, &s.Duration, &s.Price)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *ServiceRepository) GetServiceById(id int) (models.Service, error) {
	var s models.Service

	err := r.db.QueryRow(`
        SELECT id, company_id, name, description, duration, price
        FROM services WHERE id = ?`, id).
		Scan(&s.ID, &s.CompanyID, &s.Name, &s.Description, &s.Duration, &s.Price)
	if err != nil {
		return models.Service{}, err
	}

	return s, nil
}

func (r *ServiceRepository) DeleteServiceById(id int) error {
	res, err := r.db.Exec("DELETE FROM services WHERE id = ?", id)
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
