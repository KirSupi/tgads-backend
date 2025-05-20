package models

import "time"

type Campaign struct {
	Id            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	StatsCSVLink  string    `json:"stats_csv_link" db:"stats_csv_link"`
	BudgetCSVLink string    `json:"budget_csv_link" db:"budget_csv_link"`
	Text          string    `json:"text" db:"text"`
	ButtonText    string    `json:"button_text" db:"button_text"`
	Link          string    `json:"link" db:"link"`
	Active        bool      `json:"active" db:"active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
