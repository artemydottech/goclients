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
	GetAllClients(ctx context.Context) ([]models.Client, error)
	GetClientsByCompany(ctx context.Context, companyID int) ([]models.Client, error)
	GetClientById(ctx context.Context, id int) (models.Client, error)
	GetClientByPhone(ctx context.Context, companyID int, phone string) (models.Client, error)
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

func (s *ClientService) CreateClient(c models.Client) (int64, error) {
	if c.CompanyID <= 0 {
		return 0, models.Invalid("client must belong to a company")
	}

	if c.Name == "" {
		return 0, models.Invalid("client name cannot be empty")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return 0, models.Invalid("name is too long, keep it under 200 characters")
	}

	c.Phone = NormalizePhone(c.Phone)
	if len(c.Phone) < minPhoneDigits || len(c.Phone) > maxPhoneDigits {
		return 0, models.Invalid("phone must contain from %d to %d digits", minPhoneDigits, maxPhoneDigits)
	}

	if c.Email != "" && !strings.Contains(c.Email, "@") {
		return 0, models.Invalid("email is invalid")
	}

	if utf8.RuneCountInString(c.Comment) > 1000 {
		return 0, models.Invalid("comment is too long, keep it under 1000 characters")
	}

	existing, err := s.repo.GetClientByPhone(context.TODO(), c.CompanyID, c.Phone)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return 0, err
	}
	if err == nil {
		return 0, models.Invalid("client with phone %s already exists (id %d)", c.Phone, existing.ID)
	}

	return s.repo.Create(context.TODO(), c)
}

func (s *ClientService) GetAllClients() ([]models.Client, error) {
	return s.repo.GetAllClients(context.TODO())
}

func (s *ClientService) GetClientsByCompany(companyID int) ([]models.Client, error) {
	return s.repo.GetClientsByCompany(context.TODO(), companyID)
}

func (s *ClientService) GetClientById(id int) (models.Client, error) {
	return s.repo.GetClientById(context.TODO(), id)
}

func (s *ClientService) DeleteClientById(id int) error {
	return s.repo.DeleteClientById(context.TODO(), id)
}
