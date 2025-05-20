package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/timmbarton/utils/tracing"

	"backend/pkg/tgads"
)

type statsRepository struct {
	pg *sqlx.DB
}

func (r *statsRepository) Create(ctx context.Context, campaignId string, stats []*tgads.Stats) error {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	batch := struct {
		Datetime pq.StringArray
		Views    pq.Int64Array
		Clicks   pq.Int64Array
		Action   pq.Int64Array
		Spend    pq.StringArray
		CPM      pq.StringArray
	}{
		Datetime: make(pq.StringArray, 0, len(stats)),
		Views:    make(pq.Int64Array, 0, len(stats)),
		Clicks:   make(pq.Int64Array, 0, len(stats)),
		Action:   make(pq.Int64Array, 0, len(stats)),
		Spend:    make(pq.StringArray, 0, len(stats)),
		CPM:      make(pq.StringArray, 0, len(stats)),
	}

	for _, item := range stats {
		batch.Datetime = append(batch.Datetime, item.Datetime.Format(time.DateTime))
		batch.Views = append(batch.Views, int64(item.Views))
		batch.Clicks = append(batch.Clicks, int64(item.Clicks))
		batch.Action = append(batch.Action, int64(item.Actions))
		batch.Spend = append(batch.Spend, item.Spend.String())
		batch.CPM = append(batch.CPM, item.CPM.String())
	}

	_, err := r.pg.ExecContext(
		ctx,
		queryCreateStats,
		campaignId,
		batch.Datetime,
		batch.Views,
		batch.Clicks,
		batch.Action,
		batch.Spend,
		batch.CPM,
	)
	if err != nil {
		return err
	}

	return nil
}
