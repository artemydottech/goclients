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
	DeleteServiceById(ctx context.Context, id int) error
}

type CatalogService struct {
	repo ServiceRepo
}

func NewCatalogService(repo ServiceRepo) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) CreateService(ctx context.Context, item models.Service) (int64, error) {
	if item.CompanyID <= 0 {
		return 0, models.Invalid("service must belong to a company")
	}

	if item.Name == "" {
		return 0, models.Invalid("service name cannot be empty")
	}

	if utf8.RuneCountInString(item.Name) > 200 {
		return 0, models.Invalid("service name is too long, keep it under 200 characters")
	}

	if utf8.RuneCountInString(item.Description) > 500 {
		return 0, models.Invalid("description is too long, keep it under 500 characters")
	}

	if item.Duration <= 0 {
		return 0, models.Invalid("service duration must be greater than zero")
	}

	if item.Duration > maxServiceDuration {
		return 0, models.Invalid("duration cannot exceed 24 hours")
	}

	if item.Price < 0 {
		return 0, models.Invalid("price cannot be negative")
	}

	return s.repo.Create(ctx, item)
}

func (s *CatalogService) GetAllServices() ([]models.Service, error) {
	return s.repo.GetAllServices(context.TODO())
}

func (s *CatalogService) GetServicesByCompany(companyID int) ([]models.Service, error) {
	return s.repo.GetServicesByCompany(context.TODO(), companyID)
}

func (s *CatalogService) GetServiceById(id int) (models.Service, error) {
	return s.repo.GetServiceById(context.TODO(), id)
}

func (s *CatalogService) DeleteServiceById(id int) error {
	return s.repo.DeleteServiceById(context.TODO(), id)
}
