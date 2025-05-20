package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/timmbarton/utils/tracing"

	"backend/internal/models"
)

type campaignsRepository struct {
	pg *sqlx.DB
}

func (r *campaignsRepository) Create(ctx context.Context, c models.Campaign) error {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	_, err := r.pg.ExecContext(
		ctx,
		queryCreateCampaign,
		c.Id,
		c.Name,
		c.StatsCSVLink,
		c.BudgetCSVLink,
		c.Text,
		c.ButtonText,
		c.Link,
		c.Active,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *campaignsRepository) Fetch(ctx context.Context) (res []*models.Campaign, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	res = make([]*models.Campaign, 0)

	err = r.pg.SelectContext(ctx, &res, queryFetchCampaigns)
	if err != nil {
		return res, err
	}

	return res, nil
}
