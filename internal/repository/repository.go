package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"backend/internal/models"
	"backend/pkg/tgads"
)

type Repositories struct {
	Campaigns CampaignsRepository
	Stats     StatsRepository
}

func New(pg *sqlx.DB) *Repositories {
	return &Repositories{
		Campaigns: &campaignsRepository{
			pg: pg,
		},
		Stats: &statsRepository{
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
