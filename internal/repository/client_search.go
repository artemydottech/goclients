package repository

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
	"strings"
)

func (r *ClientRepository) SearchClients(ctx context.Context, filter models.ClientSearch) ([]models.Client, error) {
	query := `SELECT id, company_id, name, phone, email, comment FROM clients WHERE company_id = ?`
	args := []any{filter.CompanyID}
	if filter.Phone != "" {
		query += " AND phone = ?"
		args = append(args, filter.Phone)
	}
	if filter.Query != "" {
		query += " AND (instr(lower(name), ?) > 0 OR instr(phone, ?) > 0 OR instr(lower(email), ?) > 0)"
		q := strings.ToLower(filter.Query)
		args = append(args, q, q, q)
	}
	query, args = paginate(query+" ORDER BY id", args, filter.Page)
	return r.queryClients(ctx, query, args...)
}
