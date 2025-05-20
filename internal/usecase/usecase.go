package usecase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/timmbarton/layout/lifecycle"
	"github.com/timmbarton/utils/tracing"

	"github.com/robfig/cron/v3"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/tgads"
)

type UseCase interface {
	lifecycle.Lifecycle

	CreateCampaign(ctx context.Context, req CreateCampaignRequest) error
	FetchCampaigns(ctx context.Context) (res []*models.Campaign, err error)

	RefreshStats()
}

type Config struct {
	RefreshStatsLoadingWorkersCount int `validate:"min=1,max=10"`
}

func New(cfg Config, r *repository.Repositories, tgads *tgads.Client) UseCase {
	return &useCase{
		cfg:   cfg,
		r:     r,
		tgads: tgads,
		c:     cron.New(),
	}
}

type useCase struct {
	cfg   Config
	r     *repository.Repositories
	tgads *tgads.Client
	c     *cron.Cron
}

func (uc *useCase) Start(_ context.Context) error {
	_, err := uc.c.AddFunc(fmt.Sprintf("%d * * * *", time.Now().Minute()+1), uc.RefreshStats) // "55 * * * *"
	if err != nil {
		return err
	}

	uc.c.Start()

	return nil
}
func (uc *useCase) Stop(_ context.Context) error {
	uc.c.Stop()

	return nil
}
func (uc *useCase) GetName() string { return "Use Case" }

func (uc *useCase) RefreshStats() {
	ctx := context.Background()

	cmps, err := uc.r.Campaigns.Fetch(ctx)
	if err != nil {
		log.Println(err)
	}

	ch := make(chan *models.Campaign)
	wg := &sync.WaitGroup{}

	wg.Add(uc.cfg.RefreshStatsLoadingWorkersCount)

	go func() {
		for _, cmp := range cmps {
			ch <- cmp
		}

		close(ch)
	}()

	for range uc.cfg.RefreshStatsLoadingWorkersCount {
		go func() {
			defer wg.Done()

			stats := make([]*tgads.Stats, 0)
			rawCmp := tgads.Campaign{}

			for cmp := range ch {
				rawCmp, err = uc.tgads.GetCampaign(ctx, tgads.GetCampaignShareLink(cmp.Id))
				if err != nil {
					log.Println(err)
					continue
				}

				stats, err = uc.tgads.GetStats(ctx, rawCmp.StatsCSVLink, rawCmp.BudgetCSVLink)
				if err != nil {
					log.Println(err)
					continue
				}

				err = uc.r.Stats.Create(ctx, cmp.Id, stats)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}()
	}

	wg.Wait()
}

type CreateCampaignRequest struct {
	Link string `json:"link" validate:"required"`
	Name string `json:"name"`
}

func (uc *useCase) CreateCampaign(ctx context.Context, req CreateCampaignRequest) error {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	raw, err := uc.tgads.GetCampaign(ctx, req.Link)
	if err != nil {
		return err
	}

	c := models.Campaign{
		Id:            raw.Id,
		Name:          req.Name,
		StatsCSVLink:  raw.StatsCSVLink,
		BudgetCSVLink: raw.BudgetCSVLink,
		Text:          raw.Text,
		ButtonText:    raw.ButtonText,
		Link:          raw.Link,
		Active:        raw.Active,
	}

	err = uc.r.Campaigns.Create(ctx, c)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) FetchCampaigns(ctx context.Context) (res []*models.Campaign, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	res, err = uc.r.Campaigns.Fetch(ctx)
	if err != nil {
		return res, err
	}

	return res, nil
}
