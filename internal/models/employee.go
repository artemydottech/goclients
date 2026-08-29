package models

type Employee struct {
	ID        int    `json:"id"`
	CompanyID int    `json:"company_id"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Position  string `json:"position"`
	Avatar    string `json:"avatar"`
}
