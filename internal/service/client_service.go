package service

import (
	"database/sql"
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
	Create(c models.Client) (int64, error)
	GetAllClients() ([]models.Client, error)
	GetClientsByCompany(companyID int) ([]models.Client, error)
	GetClientById(id int) (models.Client, error)
	GetClientByPhone(companyID int, phone string) (models.Client, error)
	DeleteClientById(id int) error
}

type ClientService struct {
	repo ClientRepo
}

func NewClientService(repo ClientRepo) *ClientService {
	return &ClientService{repo: repo}
}

// NormalizePhone оставляет от номера только цифры: «+7 (999) 123-45-67» и
// «79991234567» — один и тот же человек, а уникальность считает база.
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
		return 0, models.Invalid("Клиент должен принадлежать компании!")
	}

	if c.Name == "" {
		return 0, models.Invalid("Имя клиента не может быть пустым!")
	}

	if utf8.RuneCountInString(c.Name) > 200 {
		return 0, models.Invalid("Имя слишком длинное! Не превышайте 200 символов")
	}

	c.Phone = NormalizePhone(c.Phone)
	if len(c.Phone) < minPhoneDigits || len(c.Phone) > maxPhoneDigits {
		return 0, models.Invalid("Телефон должен содержать от %d до %d цифр!", minPhoneDigits, maxPhoneDigits)
	}

	if c.Email != "" && !strings.Contains(c.Email, "@") {
		return 0, models.Invalid("Почта указана некорректно!")
	}

	if utf8.RuneCountInString(c.Comment) > 1000 {
		return 0, models.Invalid("Комментарий слишком длинный! Не превышайте 1000 символов")
	}

	existing, err := s.repo.GetClientByPhone(c.CompanyID, c.Phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if err == nil {
		return 0, models.Invalid("Клиент с телефоном %s уже есть в базе (id %d)!", c.Phone, existing.ID)
	}

	return s.repo.Create(c)
}

func (s *ClientService) GetAllClients() ([]models.Client, error) {
	return s.repo.GetAllClients()
}

func (s *ClientService) GetClientsByCompany(companyID int) ([]models.Client, error) {
	return s.repo.GetClientsByCompany(companyID)
}

func (s *ClientService) GetClientById(id int) (models.Client, error) {
	return s.repo.GetClientById(id)
}

func (s *ClientService) DeleteClientById(id int) error {
	return s.repo.DeleteClientById(id)
}
