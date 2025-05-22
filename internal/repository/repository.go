package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/timmbarton/utils/types/dates"

	"backend/internal/models"
	"backend/pkg/tgads"
)

type Repositories struct {
	Campaigns CampaignsRepository
	Stats     StatsRepository
	Rates     RatesRepository
}

func New(pg *sqlx.DB) *Repositories {
	return &Repositories{
		Campaigns: &campaignsRepository{
			pg: pg,
		},
		Stats: &statsRepository{
			pg: pg,
		},
		Rates: &ratesRepository{
			pg: pg,
		},
	}
}

type CampaignsRepository interface {
	Create(ctx context.Context, c models.Campaign) error
	Fetch(ctx context.Context) (res []*models.Campaign, err error)
}

type StatsRepository interface {
	Create(ctx context.Context, campaignId string, stats []*tgads.Stats) error
}

type RatesRepository interface {
	Create(ctx context.Context, date dates.Date, rate decimal.Decimal) error
}
