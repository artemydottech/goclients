package models

import "time"

type ReportQuery struct {
	CompanyID int
	From      time.Time
	To        time.Time
}

func (q ReportQuery) Validate() error {
	if q.CompanyID <= 0 {
		return Invalid("company_id must be positive")
	}
	if q.From.IsZero() || q.To.IsZero() || !q.To.After(q.From) {
		return Invalid("from and to must define a positive period")
	}
	if q.To.Sub(q.From) > 366*24*time.Hour {
		return Invalid("report period cannot exceed 366 days")
	}
	return nil
}

type ReportSummary struct {
	Appointments    int     `json:"appointments"`
	Active          int     `json:"active"`
	Completed       int     `json:"completed"`
	Cancelled       int     `json:"cancelled"`
	NoShows         int     `json:"no_shows"`
	UniqueClients   int     `json:"unique_clients"`
	Revenue         float64 `json:"revenue"`
	ExpectedRevenue float64 `json:"expected_revenue"`
}
type ReportGroup struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Appointments int     `json:"appointments"`
	Completed    int     `json:"completed"`
	Revenue      float64 `json:"revenue"`
}
type CompanyReport struct {
	CompanyID int           `json:"company_id"`
	From      time.Time     `json:"from"`
	To        time.Time     `json:"to"`
	Summary   ReportSummary `json:"summary"`
	Services  []ReportGroup `json:"services"`
	Employees []ReportGroup `json:"employees"`
}
