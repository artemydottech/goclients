package models

type Service struct {
	ID          int     `json:"id"`
	CompanyID   int     `json:"company_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	Price       float64 `json:"price"`
}
