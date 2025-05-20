package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/timmbarton/layout/components/httpserver"
	"github.com/timmbarton/layout/lifecycle"

	"backend/internal/usecase"
)

func New(cfg httpserver.Config, uc usecase.UseCase) lifecycle.Lifecycle {
	s := &httpserver.DefaultServer{}

	h := &handler{
		uc: uc,
		v:  validator.New(),
	}

	s.Init(cfg, h.bind)

	return s
}
