package repository

import (
	"context"
	"strings"

	"github.com/artemydottech/goclients/internal/models"
)

func (r *ClientRepository) SearchClients(ctx context.Context, filter models.ClientSearch) ([]models.Client, error) {
	query := `SELECT id, company_id, name, phone, email, comment FROM clients WHERE company_id = ?`
	args := []any{filter.CompanyID}
	if filter.Phone != "" {
		query += " AND phone = ?"
		args = append(args, filter.Phone)
	}
	if filter.Tag != "" {
		query += " AND EXISTS(SELECT 1 FROM client_tags t WHERE t.client_id=clients.id AND t.tag=?)"
		args = append(args, filter.Tag)
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
	phoneQuery := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, filter.Query)
	for _, c := range items {
		if strings.Contains(strings.ToLower(c.Name), q) || (phoneQuery != "" && strings.Contains(c.Phone, phoneQuery)) || strings.Contains(strings.ToLower(c.Email), q) {
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
