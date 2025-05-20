package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/timmbarton/response"
	"github.com/timmbarton/utils/tracing"

	"backend/internal/usecase"
	"backend/pkg/errlist"
)

func (h *handler) campaignsGet(c *fiber.Ctx) error {
	ctx, span := tracing.NewSpan(c.UserContext())
	defer span.End()
	c.SetUserContext(ctx)

	res, err := h.uc.FetchCampaigns(ctx)
	if err != nil {
		return err
	}

	return response.OkWithData(c, res)
}

func (h *handler) campaignsPost(c *fiber.Ctx) error {
	ctx, span := tracing.NewSpan(c.UserContext())
	defer span.End()
	c.SetUserContext(ctx)

	req := usecase.CreateCampaignRequest{}

	err := c.BodyParser(&req)
	if err != nil {
		return errlist.ErrBadRequest
	}

	err = h.v.Struct(req)
	if err != nil {
		return errlist.ErrBadRequest
	}

	err = h.uc.CreateCampaign(ctx, req)
	if err != nil {
		return err
	}

	return response.Ok(c)
}
