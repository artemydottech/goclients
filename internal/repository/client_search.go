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
	query += " ORDER BY id"
	if filter.Query == "" {
		query, args = paginate(query, args, filter.Page)
		return r.queryClients(ctx, query, args...)
	}
	items, err := r.queryClients(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	found := []models.Client{}
	q := strings.ToLower(filter.Query)
	for _, c := range items {
		if strings.Contains(strings.ToLower(c.Name), q) || strings.Contains(c.Phone, q) || strings.Contains(strings.ToLower(c.Email), q) {
			found = append(found, c)
		}
	}
	start := min(filter.Page.Offset, len(found))
	end := len(found)
	if filter.Page.Limit > 0 {
		end = min(start+filter.Page.Limit, len(found))
	}
	return found[start:end], nil
}
