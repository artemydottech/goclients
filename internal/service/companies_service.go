package service

import (
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyRepo interface {
	Create(c models.Company) (int64, error)
	GetAllCompanies() ([]models.Company, error)
	GetCompanyById(id int) (models.Company, error)
	DeleteCompanyById(id int) error
}

type CompanyService struct {
	repo CompanyRepo
}

func NewCompanyService(repo CompanyRepo) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) CreateCompany(c models.Company) (int64, error) {
	if c.Name == "" {
		return 0, models.Invalid("Название компании не может быть пустым!")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return 0, models.Invalid("Название компании слишком длинное! Не превышайте 200 символов")
	}

	if c.Address != "" && utf8.RuneCountInString(c.Address) > 500 {
		return 0, models.Invalid("Адрес слишком длинный! Не превышайте 500 символов")
	}

	if c.Site != "" && utf8.RuneCountInString(c.Site) > 500 {
		return 0, models.Invalid("Сайт слишком длинный! Не превышайте 500 символов")
	}

	if c.Logo != "" && utf8.RuneCountInString(c.Logo) > 500 {
		return 0, models.Invalid("Ссылка на логотип слишком длинная! Не превышайте 500 символов")
	}

	if err := c.Socials.Validate(); err != nil {
		return 0, err
	}

	id, err := s.repo.Create(c)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *CompanyService) GetAllCompanies() ([]models.Company, error) {
	return s.repo.GetAllCompanies()
}

func (s *CompanyService) GetCompanyById(id int) (models.Company, error) {
	return s.repo.GetCompanyById(id)
}

func (s *CompanyService) DeleteCompanyById(id int) error {
	return s.repo.DeleteCompanyById(id)
}
