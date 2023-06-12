package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type NotificationHandler struct {
	Service *services.NotificationStorage
}

func NewNotificationHandler(service *services.NotificationStorage) *NotificationHandler {
	return &NotificationHandler{
		Service: service,
	}
}

func (h *NotificationHandler) CreateNotificationPair(ctx *fiber.Ctx)  error {
	var req services.NotificationRequest

	err := ctx.BodyParser(&req)

	if err != nil {
		return ctx.Status(422).JSON(&fiber.Map{
			"message" : "Malformed request body " + err.Error(),
		})
	}

	result, err := h.Service.CreateNotificationPair(req)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message" : "Internal server error " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(result)
}


func (h *NotificationHandler) FetchDeviceTokenForUser(ctx *fiber.Ctx)  error {
	userId := ctx.Params("userId")

	if userId == "" {
		return ctx.Status(422).JSON(&fiber.Map{
			"message" : "Malformed request body ",
		})
	}

	result, err := h.Service.RetriveDeviceTokenForUser(userId)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message" : "Internal server error " + err.Error(),
		})
	}

	return ctx.Status(200).JSON(&fiber.Map{
		"deviceId" : result,
	})
}