package app

import (
	"github.com/timmbarton/layout/components/postgresconn"
	"github.com/timmbarton/layout/components/tracingconn"
	"github.com/timmbarton/layout/executor"
	"github.com/timmbarton/layout/template"

	"backend/internal/config"
	"backend/internal/delivery/http"
	"backend/internal/repository"
	"backend/internal/usecase"
	"backend/pkg/coingecko"
	"backend/pkg/tgads"
)

func New(cfg config.Config) (executor.App, error) {
	a := &template.App{}

	pg, err := postgresconn.New(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	r := repository.New(pg.DB())
	uc := usecase.New(cfg.UseCase, r, tgads.New(), coingecko.New(cfg.CoinGeckoApiKey))
	httpServer := http.New(cfg.HTTP, uc)

	a.AddComponents(
		tracingconn.New(cfg.Tracing),
		pg,
		uc,
		httpServer,
	)

	return a, nil
}
