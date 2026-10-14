package service

import (
	"context"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyRepo interface {
	Create(ctx context.Context, c models.Company) (int64, error)
	GetAllCompanies(ctx context.Context) ([]models.Company, error)
	GetCompanyById(ctx context.Context, id int) (models.Company, error)
	DeleteCompanyById(ctx context.Context, id int) error
}

type CompanyService struct {
	repo CompanyRepo
}

func NewCompanyService(repo CompanyRepo) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) CreateCompany(c models.Company) (int64, error) {
	if c.Name == "" {
		return 0, models.Invalid("company name cannot be empty")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return 0, models.Invalid("company name is too long, keep it under 200 characters")
	}

	if c.Address != "" && utf8.RuneCountInString(c.Address) > 500 {
		return 0, models.Invalid("address is too long, keep it under 500 characters")
	}

	if c.Site != "" && utf8.RuneCountInString(c.Site) > 500 {
		return 0, models.Invalid("website is too long, keep it under 500 characters")
	}

	if c.Logo != "" && utf8.RuneCountInString(c.Logo) > 500 {
		return 0, models.Invalid("logo URL is too long, keep it under 500 characters")
	}

	if err := c.Socials.Validate(); err != nil {
		return 0, err
	}

	if c.Timezone == "" {
		c.Timezone = "UTC"
	}
	if _, err := c.Location(); err != nil {
		return 0, models.Invalid("unknown timezone %s, use an IANA name such as Asia/Yekaterinburg", c.Timezone)
	}

	id, err := s.repo.Create(context.TODO(), c)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *CompanyService) GetAllCompanies() ([]models.Company, error) {
	return s.repo.GetAllCompanies(context.TODO())
}

func (s *CompanyService) GetCompanyById(id int) (models.Company, error) {
	return s.repo.GetCompanyById(context.TODO(), id)
}

func (s *CompanyService) DeleteCompanyById(id int) error {
	return s.repo.DeleteCompanyById(context.TODO(), id)
}
