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

type APNSNotificationRequest struct {
	Message string `json:"message"`
}



func (h *NotificationHandler) SendAPNSNotification(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	var req APNSNotificationRequest

	err := ctx.BodyParser(&req)

	if err != nil {
		return ctx.Status(422).JSON(&fiber.Map{
			"message" : "Malformed request body " + err.Error(),
		})
	}

	if len(userId) == 0 {
		return ctx.Status(400).JSON(&fiber.Map{
			"message" : "PAram not given..",
		})
	}

	res, err := h.Service.CreateAPNSNotification(req.Message, userId)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message" : "Internal server error " + err.Error(),
		})
	}

	if res.Sent() {
		return ctx.Status(200).JSON(&fiber.Map{
			"message" : "Notification sent successfully",
		})
	} else {
		return ctx.Status(500).JSON(&fiber.Map{
			"message" : "Notification not sent " + res.Reason,
		})
	}
}