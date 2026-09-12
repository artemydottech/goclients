package service

import (
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

// maxServiceDuration — сутки: запись длиннее рабочего дня не бронируется.
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
		return 0, models.Invalid("Услуга должна принадлежать компании!")
	}

	if item.Name == "" {
		return 0, models.Invalid("Название услуги не может быть пустым!")
	}

	if utf8.RuneCountInString(item.Name) > 200 {
		return 0, models.Invalid("Название слишком длинное! Не превышайте 200 символов")
	}

	if utf8.RuneCountInString(item.Description) > 500 {
		return 0, models.Invalid("Описание слишком длинное! Не превышайте 500 символов")
	}

	if item.Duration <= 0 {
		return 0, models.Invalid("Длительность услуги должна быть больше нуля!")
	}

	if item.Duration > maxServiceDuration {
		return 0, models.Invalid("Длительность не может превышать 24 часа!")
	}

	if item.Price < 0 {
		return 0, models.Invalid("Цена не может быть отрицательной!")
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
