package service

import (
	"context"
	"github.com/artemydottech/goclients/internal/models"
)

type ClientTagRepo interface {
	GetClientTags(context.Context, int) ([]string, error)
	ReplaceClientTags(context.Context, int, []string) error
}
type ClientTagService struct {
	repo    ClientTagRepo
	clients ClientLookup
}

func NewClientTagService(r ClientTagRepo, c ClientLookup) *ClientTagService {
	return &ClientTagService{repo: r, clients: c}
}
func (s *ClientTagService) Get(ctx context.Context, id int) ([]string, error) {
	if id <= 0 {
		return nil, models.Invalid("invalid client id")
	}
	if _, err := s.clients.GetClientById(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.GetClientTags(ctx, id)
}
func (s *ClientTagService) Replace(ctx context.Context, id int, tags []string) error {
	if id <= 0 {
		return models.Invalid("invalid client id")
	}
	normalized, err := models.NormalizeClientTags(tags)
	if err != nil {
		return err
	}
	if _, err := s.clients.GetClientById(ctx, id); err != nil {
		return err
	}
	return s.repo.ReplaceClientTags(ctx, id, normalized)
}
