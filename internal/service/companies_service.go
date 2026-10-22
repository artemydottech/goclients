package service

import (
	"context"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyRepo interface {
	Create(ctx context.Context, c models.Company) (int64, error)
	GetAllCompanies(ctx context.Context, page models.Page) ([]models.Company, error)
	GetCompanyById(ctx context.Context, id int) (models.Company, error)
	UpdateCompany(ctx context.Context, c models.Company) error
	DeleteCompanyById(ctx context.Context, id int) error
}

type CompanyService struct {
	repo CompanyRepo
}

func NewCompanyService(repo CompanyRepo) *CompanyService {
	return &CompanyService{repo: repo}
}

func validateCompany(c models.Company) (models.Company, error) {
	if c.Name == "" {
		return c, models.Invalid("company name cannot be empty")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return c, models.Invalid("company name is too long, keep it under 200 characters")
	}

	if c.Address != "" && utf8.RuneCountInString(c.Address) > 500 {
		return c, models.Invalid("address is too long, keep it under 500 characters")
	}

	if c.Site != "" && utf8.RuneCountInString(c.Site) > 500 {
		return c, models.Invalid("website is too long, keep it under 500 characters")
	}

	if c.Logo != "" && utf8.RuneCountInString(c.Logo) > 500 {
		return c, models.Invalid("logo URL is too long, keep it under 500 characters")
	}

	if err := c.Socials.Validate(); err != nil {
		return c, err
	}

	if c.Timezone == "" {
		c.Timezone = "UTC"
	}
	if _, err := c.Location(); err != nil {
		return c, models.Invalid("unknown timezone %s, use an IANA name such as Asia/Yekaterinburg", c.Timezone)
	}

	return c, nil
}

func (s *CompanyService) CreateCompany(ctx context.Context, c models.Company) (int64, error) {
	c, err := validateCompany(c)
	if err != nil {
		return 0, err
	}

	id, err := s.repo.Create(ctx, c)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id int, c models.Company) error {
	c, err := validateCompany(c)
	if err != nil {
		return err
	}

	c.ID = id

	return s.repo.UpdateCompany(ctx, c)
}

func (s *CompanyService) GetAllCompanies(ctx context.Context, page models.Page) ([]models.Company, error) {
	return s.repo.GetAllCompanies(ctx, page)
}

func (s *CompanyService) GetCompanyById(ctx context.Context, id int) (models.Company, error) {
	return s.repo.GetCompanyById(ctx, id)
}

func (s *CompanyService) DeleteCompanyById(ctx context.Context, id int) error {
	return s.repo.DeleteCompanyById(ctx, id)
}
