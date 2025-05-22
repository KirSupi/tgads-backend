package usecase

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/timmbarton/layout/lifecycle"
	"github.com/timmbarton/utils/tracing"
	"github.com/timmbarton/utils/types/dates"

	"github.com/robfig/cron/v3"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/coingecko"
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

func New(cfg Config, r *repository.Repositories, tgads *tgads.Client, cg *coingecko.Client) UseCase {
	return &useCase{
		cfg:   cfg,
		r:     r,
		tgads: tgads,
		cg:    cg,
		c:     cron.New(),
	}
}

type useCase struct {
	cfg   Config
	r     *repository.Repositories
	tgads *tgads.Client
	cg    *coingecko.Client
	c     *cron.Cron
}

func (uc *useCase) Start(_ context.Context) error {
	_, err := uc.c.AddFunc("45 * * * *", uc.LoadRates)
	if err != nil {
		return err
	}

	_, err = uc.c.AddFunc("55 * * * *", uc.RefreshStats)
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

// LoadRates подгружает курс TON к USD
func (uc *useCase) LoadRates() {
	ctx := context.Background()

	yesterdayDate := dates.Date(time.Now().AddDate(0, 0, -1))
	todayDate := dates.Date(time.Now())

	yesterdayRate, err := uc.cg.GetTonRate(ctx, yesterdayDate)
	if err != nil {
		log.Println(err)
		return
	}

	err = uc.r.Rates.Create(ctx, yesterdayDate, yesterdayRate)
	if err != nil {
		log.Println(err)
		return
	}

	todayRate, err := uc.cg.GetTonRate(ctx, todayDate)
	if err != nil {
		log.Println(err)
		return
	}

	err = uc.r.Rates.Create(ctx, todayDate, todayRate)
	if err != nil {
		log.Println(err)
		return
	}
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
