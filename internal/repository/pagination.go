package repository

import "github.com/artemydottech/goclients/internal/models"

func paginate(query string, args []any, page models.Page) (string, []any) {
	switch {
	case page.Limit > 0:
		return query + " LIMIT ? OFFSET ?", append(args, page.Limit, page.Offset)
	case page.Offset > 0:
		return query + " LIMIT -1 OFFSET ?", append(args, page.Offset)
	}

	return query, args
}
