package repository

import (
	"database/sql"
	"log"

	"github.com/artemydottech/goclients/internal/models"
)

type EmployeeRepository struct {
	db *sql.DB
}

func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(e models.Employee) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO employees (company_id, name, surname, position, avatar) VALUES (?, ?, ?, ?, ?)",
		e.CompanyID, e.Name, e.Surname, e.Position, e.Avatar,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *EmployeeRepository) GetAllEmployees() ([]models.Employee, error) {
	rows, err := r.db.Query(`
        SELECT id, company_id, name, surname, position, avatar
        FROM employees`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := []models.Employee{}
	for rows.Next() {
		var e models.Employee
		err := rows.Scan(&e.ID, &e.CompanyID, &e.Name, &e.Surname, &e.Position, &e.Avatar)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *EmployeeRepository) GetEmployeeById(id int) (models.Employee, error) {
	var e models.Employee

	err := r.db.QueryRow(`
        SELECT id, company_id, name, surname, position, avatar
        FROM employees WHERE id = ?`, id).
		Scan(&e.ID, &e.CompanyID, &e.Name, &e.Surname, &e.Position, &e.Avatar)
	if err != nil {
		return models.Employee{}, err
	}

	return e, nil
}

func (r *EmployeeRepository) DeleteEmployeeById(id int) error {
	res, err := r.db.Exec("DELETE FROM employees WHERE id = ?", id)
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

func (r *EmployeeRepository) TestRows() {
	rows, _ := r.db.Query("SELECT id, name, position FROM employees")
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var name, position string
		rows.Scan(&id, &name, &position)
		log.Printf("Сотрудник #%d: ID=%d Name=%s position=%s", count, id, name, position)
		count++
	}
	log.Printf("Всего сотрудников: %d", count)
}
