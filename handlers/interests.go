package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type InterestHandler struct {
	Service *services.InterestStorage
}

func NewInterestHandler(service *services.InterestStorage) *InterestHandler {
	return &InterestHandler{
		Service: service,
	}
}


func (h *InterestHandler) CreateBatchInterests(ctx *fiber.Ctx) error {
	var interests []services.Interest

	err := ctx.BodyParser(&interests)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	return ctx.Status(201).JSON(h.Service.BatchCreateInterests(interests))
}