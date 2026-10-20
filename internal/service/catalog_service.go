package service

import (
	"context"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

const maxServiceDuration = 24 * 60

type ServiceRepo interface {
	Create(ctx context.Context, s models.Service) (int64, error)
	GetAllServices(ctx context.Context) ([]models.Service, error)
	GetServicesByCompany(ctx context.Context, companyID int) ([]models.Service, error)
	GetServiceById(ctx context.Context, id int) (models.Service, error)
	UpdateService(ctx context.Context, s models.Service) error
	DeleteServiceById(ctx context.Context, id int) error
}

type CatalogService struct {
	repo ServiceRepo
}

func NewCatalogService(repo ServiceRepo) *CatalogService {
	return &CatalogService{repo: repo}
}

func validateService(item models.Service) error {
	if item.Name == "" {
		return models.Invalid("service name cannot be empty")
	}

	if utf8.RuneCountInString(item.Name) > 200 {
		return models.Invalid("service name is too long, keep it under 200 characters")
	}

	if utf8.RuneCountInString(item.Description) > 500 {
		return models.Invalid("description is too long, keep it under 500 characters")
	}

	if item.Duration <= 0 {
		return models.Invalid("service duration must be greater than zero")
	}

	if item.Duration > maxServiceDuration {
		return models.Invalid("duration cannot exceed 24 hours")
	}

	if item.Price < 0 {
		return models.Invalid("price cannot be negative")
	}

	return nil
}

func (s *CatalogService) CreateService(ctx context.Context, item models.Service) (int64, error) {
	if item.CompanyID <= 0 {
		return 0, models.Invalid("service must belong to a company")
	}

	if err := validateService(item); err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, item)
}

func (s *CatalogService) UpdateService(ctx context.Context, id int, item models.Service) error {
	if err := validateService(item); err != nil {
		return err
	}

	item.ID = id

	return s.repo.UpdateService(ctx, item)
}

func (s *CatalogService) GetAllServices(ctx context.Context) ([]models.Service, error) {
	return s.repo.GetAllServices(ctx)
}

func (s *CatalogService) GetServicesByCompany(ctx context.Context, companyID int) ([]models.Service, error) {
	return s.repo.GetServicesByCompany(ctx, companyID)
}

func (s *CatalogService) GetServiceById(ctx context.Context, id int) (models.Service, error) {
	return s.repo.GetServiceById(ctx, id)
}

func (s *CatalogService) DeleteServiceById(ctx context.Context, id int) error {
	return s.repo.DeleteServiceById(ctx, id)
}
