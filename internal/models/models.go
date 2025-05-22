package models

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/timmbarton/utils/types/dates"
)

// Campaign описывает информацию о добавленной РК
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

// Stats описывает статистику по РК за определённую дату
type Stats struct {
	CampaignId string          `json:"campaign_id" db:"campaign_id"`
	Date       dates.Date      `json:"date" db:"date"`
	Views      int             `json:"views" db:"views"`
	Clicks     int             `json:"clicks" db:"clicks"`
	Actions    int             `json:"actions" db:"actions"`
	Spend      decimal.Decimal `json:"spend" db:"spend"`
	Cpm        decimal.Decimal `json:"cpm" db:"cpm"`
}

// Rate описывает курс TON к USD
type Rate struct {
	Date dates.Date      `json:"date" db:"date"`
	Rate decimal.Decimal `json:"rate" db:"rate"`
}
