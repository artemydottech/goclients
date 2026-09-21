package service

import (
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

const maxServiceDuration = 24 * 60

type ServiceRepo interface {
	Create(s models.Service) (int64, error)
	GetAllServices() ([]models.Service, error)
	GetServicesByCompany(companyID int) ([]models.Service, error)
	GetServiceById(id int) (models.Service, error)
	DeleteServiceById(id int) error
}

type CatalogService struct {
	repo ServiceRepo
}

func NewCatalogService(repo ServiceRepo) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) CreateService(item models.Service) (int64, error) {
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

	return s.repo.Create(item)
}

func (s *CatalogService) GetAllServices() ([]models.Service, error) {
	return s.repo.GetAllServices()
}

func (s *CatalogService) GetServicesByCompany(companyID int) ([]models.Service, error) {
	return s.repo.GetServicesByCompany(companyID)
}

func (s *CatalogService) GetServiceById(id int) (models.Service, error) {
	return s.repo.GetServiceById(id)
}

func (s *CatalogService) DeleteServiceById(id int) error {
	return s.repo.DeleteServiceById(id)
}
