package models

// Client — клиент конкретной компании. Телефон опознаёт человека: по нему
// салон находит карточку, поэтому он уникален в пределах компании.
type Client struct {
	ID        int    `json:"id"`
	CompanyID int    `json:"company_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Comment   string `json:"comment"`
}
