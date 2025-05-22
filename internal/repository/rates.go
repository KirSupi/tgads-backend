package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/timmbarton/utils/tracing"
	"github.com/timmbarton/utils/types/dates"
)

type ratesRepository struct {
	pg *sqlx.DB
}

func (r *ratesRepository) Create(ctx context.Context, date dates.Date, rate decimal.Decimal) error {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	_, err := r.pg.ExecContext(
		ctx,
		queryCreateRate,
		date,
		rate,
	)
	if err != nil {
		return err
	}

	return nil
}
