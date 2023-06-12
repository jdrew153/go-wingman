package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type WMNotificationHandler struct {
	Service *services.WMNotificationStorage
}

func NewWMNotificationHandler(service *services.WMNotificationStorage) *WMNotificationHandler {
	return &WMNotificationHandler{
		Service: service,
	}
}

func (h *WMNotificationHandler) CreateWMNotification(ctx *fiber.Ctx) error {
	var notification services.WMNotificationRequest

	if err := ctx.BodyParser(&notification); err != nil {
		return ctx.Status(422).JSON(&fiber.Map{
			"message" : "Malformed request body" + err.Error(),
		})
	}

	result, err := h.Service.CreateNotification(notification)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message" : "Internal server error" + err.Error(),
		})
	}

	return ctx.Status(201).JSON(result)
}