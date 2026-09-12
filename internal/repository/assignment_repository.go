package repository

import (
	"database/sql"

	"github.com/artemydottech/goclients/internal/models"
)

type AssignmentRepository struct {
	db *sql.DB
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// SetEmployeeServices заменяет весь набор услуг сотрудника целиком: половина
// записанного набора хуже, чем ни одной, поэтому всё в одной транзакции.
func (r *AssignmentRepository) SetEmployeeServices(employeeID int, serviceIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM employee_services WHERE employee_id = ?", employeeID); err != nil {
		return err
	}

	statement, err := tx.Prepare("INSERT INTO employee_services (employee_id, service_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	for _, serviceID := range serviceIDs {
		if _, err := statement.Exec(employeeID, serviceID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AssignmentRepository) GetServicesByEmployee(employeeID int) ([]models.Service, error) {
	rows, err := r.db.Query(`
        SELECT s.id, s.company_id, s.name, s.description, s.duration, s.price
        FROM services s
        JOIN employee_services es ON es.service_id = s.id
        WHERE es.employee_id = ?
        ORDER BY s.id`, employeeID)
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

	return services, rows.Err()
}

func (r *AssignmentRepository) GetEmployeesByService(serviceID int) ([]models.Employee, error) {
	rows, err := r.db.Query(`
        SELECT e.id, e.company_id, e.name, e.surname, e.position, e.avatar
        FROM employees e
        JOIN employee_services es ON es.employee_id = e.id
        WHERE es.service_id = ?
        ORDER BY e.id`, serviceID)
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

	return employees, rows.Err()
}

func (r *AssignmentRepository) EmployeePerformsService(employeeID, serviceID int) (bool, error) {
	var exists int

	err := r.db.QueryRow(
		"SELECT 1 FROM employee_services WHERE employee_id = ? AND service_id = ?",
		employeeID, serviceID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
