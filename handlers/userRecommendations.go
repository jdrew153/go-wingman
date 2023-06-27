package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type UserRecommendationHandler struct {
	Service *services.UserRecommendationService
}

func NewUserRecommendationHandler(s *services.UserRecommendationService) *UserRecommendationHandler {
	return &UserRecommendationHandler{
		Service: s,
	}
}

func (h *UserRecommendationHandler) CreateRecommendationsForUser(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	fmt.Println(userId)

	recommendations, err := h.Service.BuildUserRecommendationModel(userId)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(recommendations)
}


func (h *UserRecommendationHandler) GetListOfRecommendedUsersContextHandler(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	fmt.Println(userId)

	recommendations, err := h.Service.GetListOfRecommendedUsersContext(userId)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(recommendations)
}