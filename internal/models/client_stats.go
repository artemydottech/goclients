package models

import "time"

type ClientStats struct {
	ClientID        int        `json:"client_id"`
	Appointments    int        `json:"appointments"`
	CompletedVisits int        `json:"completed_visits"`
	NoShows         int        `json:"no_shows"`
	Cancelled       int        `json:"cancelled"`
	Upcoming        int        `json:"upcoming"`
	TotalSpent      float64    `json:"total_spent"`
	LastVisitAt     *time.Time `json:"last_visit_at"`
	NextVisitAt     *time.Time `json:"next_visit_at"`
}
