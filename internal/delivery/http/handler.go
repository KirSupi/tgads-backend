package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"backend/internal/usecase"
)

type handler struct {
	uc usecase.UseCase
	v  *validator.Validate
}

func (h *handler) bind(r fiber.Router) {
	campaignsGroup := r.Group("/campaigns")
	{
		campaignsGroup.Get("/", h.campaignsGet)
		campaignsGroup.Post("/", h.campaignsPost)
	}
}
