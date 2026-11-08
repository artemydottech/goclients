package service

import (
	"context"
	"strings"

	"github.com/artemydottech/goclients/internal/models"
)

type ClientSearchRepo interface {
	SearchClients(context.Context, models.ClientSearch) ([]models.Client, error)
}
type ClientSearchService struct{ repo ClientSearchRepo }

func NewClientSearchService(repo ClientSearchRepo) *ClientSearchService {
	return &ClientSearchService{repo: repo}
}
func (s *ClientSearchService) SearchClients(ctx context.Context, filter models.ClientSearch) ([]models.Client, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	if filter.Phone != "" {
		filter.Phone = NormalizePhone(filter.Phone)
		if filter.Phone == "" {
			return nil, models.Invalid("invalid phone")
		}
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return s.repo.SearchClients(ctx, filter)
}
