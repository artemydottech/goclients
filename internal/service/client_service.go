package service

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/artemydottech/goclients/internal/models"
)

const (
	minPhoneDigits = 10
	maxPhoneDigits = 15
)

type ClientRepo interface {
	Create(ctx context.Context, c models.Client) (int64, error)
	GetAllClients(ctx context.Context, page models.Page) ([]models.Client, error)
	GetClientsByCompany(ctx context.Context, companyID int, page models.Page) ([]models.Client, error)
	GetClientById(ctx context.Context, id int) (models.Client, error)
	GetClientByPhone(ctx context.Context, companyID int, phone string) (models.Client, error)
	UpdateClient(ctx context.Context, c models.Client) error
	DeleteClientById(ctx context.Context, id int) error
}

type ClientService struct {
	repo ClientRepo
}

func NewClientService(repo ClientRepo) *ClientService {
	return &ClientService{repo: repo}
}

func NormalizePhone(phone string) string {
	var digits strings.Builder

	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}

	return digits.String()
}

func validateClient(c models.Client) (models.Client, error) {
	if c.Name == "" {
		return c, models.Invalid("client name cannot be empty")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return c, models.Invalid("name is too long, keep it under 200 characters")
	}

	c.Phone = NormalizePhone(c.Phone)
	if len(c.Phone) < minPhoneDigits || len(c.Phone) > maxPhoneDigits {
		return c, models.Invalid("phone must contain from %d to %d digits", minPhoneDigits, maxPhoneDigits)
	}

	if c.Email != "" && !strings.Contains(c.Email, "@") {
		return c, models.Invalid("email is invalid")
	}

	if utf8.RuneCountInString(c.Comment) > 1000 {
		return c, models.Invalid("comment is too long, keep it under 1000 characters")
	}

	return c, nil
}

func (s *ClientService) CreateClient(ctx context.Context, c models.Client) (int64, error) {
	if c.CompanyID <= 0 {
		return 0, models.Invalid("client must belong to a company")
	}

	c, err := validateClient(c)
	if err != nil {
		return 0, err
	}

	existing, err := s.repo.GetClientByPhone(ctx, c.CompanyID, c.Phone)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return 0, err
	}
	if err == nil {
		return 0, models.Invalid("client with phone %s already exists (id %d)", c.Phone, existing.ID)
	}

	return s.repo.Create(ctx, c)
}

func (s *ClientService) UpdateClient(ctx context.Context, id int, c models.Client) error {
	current, err := s.repo.GetClientById(ctx, id)
	if err != nil {
		return err
	}

	c, err = validateClient(c)
	if err != nil {
		return err
	}

	other, err := s.repo.GetClientByPhone(ctx, current.CompanyID, c.Phone)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if err == nil && other.ID != id {
		return models.Invalid("client with phone %s already exists (id %d)", c.Phone, other.ID)
	}

	c.ID = id
	c.CompanyID = current.CompanyID

	return s.repo.UpdateClient(ctx, c)
}

func (s *ClientService) GetAllClients(ctx context.Context) ([]models.Client, error) {
	return s.repo.GetAllClients(ctx, models.Page{})
}

func (s *ClientService) GetClientsByCompany(ctx context.Context, companyID int) ([]models.Client, error) {
	return s.repo.GetClientsByCompany(ctx, companyID, models.Page{})
}

func (s *ClientService) GetClientById(ctx context.Context, id int) (models.Client, error) {
	return s.repo.GetClientById(ctx, id)
}

func (s *ClientService) DeleteClientById(ctx context.Context, id int) error {
	return s.repo.DeleteClientById(ctx, id)
}
