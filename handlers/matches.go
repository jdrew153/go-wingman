package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type MatchHandler struct {
	MatchService *services.MatchStorage
	NotificationService *services.NotificationStorage
}

func NewMatchHandler(m *services.MatchStorage, n *services.NotificationStorage) *MatchHandler {
	return &MatchHandler{
		MatchService: m,
		NotificationService: n,
	}
}

func (h *MatchHandler) CreateMatch(ctx *fiber.Ctx) error {
	var request services.MatchRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	match, err := h.MatchService.CreateNewMatch(request)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(match)
}

type MatchUpdateRequest struct {
	UserIdA string `json:"userIdA"`
	UserIdB string `json:"userIdB"`
	MatchStatus string `json:"matchStatus"`
}

func (h *MatchHandler) FindAndUpDateMatchHandler(ctx *fiber.Ctx) error {
	var request MatchUpdateRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	match, err := h.MatchService.FindAndUpdateMatchStatus(request.UserIdA, request.UserIdB, request.MatchStatus)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(200).JSON(match)
}