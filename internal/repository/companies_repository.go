package repository

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(c models.Company) (int64, error) {
	socialJSON, err := json.Marshal(c.Socials)
	if err != nil {
		return 0, err
	}

	res, err := r.db.Exec(
		"INSERT INTO companies (name, address, geolocation, schedule, site, socials, logo) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.Name, c.Address, c.Geolocation, c.Schedule, c.Site, string(socialJSON), c.Logo,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *CompanyRepository) GetAllCompanies() ([]models.Company, error) {
	rows, err := r.db.Query(`
        SELECT id, name, address, geolocation, schedule, site, socials, logo 
        FROM companies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var c models.Company
		var socialJSON sql.NullString

		err := rows.Scan(&c.ID, &c.Name, &c.Address, &c.Geolocation,
			&c.Schedule, &c.Site, &socialJSON, &c.Logo)
		if err != nil {
			return nil, err
		}

		if socialJSON.Valid {
			json.Unmarshal([]byte(socialJSON.String), &c.Socials)
		}

		companies = append(companies, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return companies, nil
}

func (r *CompanyRepository) GetCompanyById(id int) (models.Company, error) {
	var c models.Company
	var socialJSON sql.NullString

	err := r.db.QueryRow(`
        SELECT id, name, address, geolocation, schedule, site, socials, logo 
        FROM companies WHERE id = ?`, id).
		Scan(&c.ID, &c.Name, &c.Address, &c.Geolocation,
			&c.Schedule, &c.Site, &socialJSON, &c.Logo)
	if err != nil {
		return models.Company{}, err
	}

	if socialJSON.Valid {
		json.Unmarshal([]byte(socialJSON.String), &c.Socials)
	}

	return c, nil
}

func (r *CompanyRepository) DeleteCompanyById(id int) error {
	res, err := r.db.Exec("DELETE FROM companies WHERE id = ?", id)
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

func (r *CompanyRepository) TestRows() {
	rows, _ := r.db.Query("SELECT id, name, socials FROM companies")
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var name, socials string
		rows.Scan(&id, &name, &socials)
		log.Printf("Компания #%d: ID=%d Name=%s Socials=%s", count, id, name, socials)
		count++
	}
	log.Printf("Всего компаний: %d", count)
}
