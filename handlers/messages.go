package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type MessageHandler struct {
	Service *services.MessageService
}

func NewMessageHandler(s *services.MessageService) *MessageHandler {
	return &MessageHandler{
		s,
	}
}

func (h *MessageHandler) SendMessageComplete(ctx *fiber.Ctx) error {
	var messageRequest services.MessageRequestModel

	err := ctx.BodyParser(&messageRequest)

	if err != nil {
		return ctx.Status(422).JSON(&fiber.Map{
			"message": "Malformed body..",
		})
	}

	fmt.Println(messageRequest)

	err = h.Service.CreateConversationKeyForUser(messageRequest.UserIdA, messageRequest.UserIdB)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message": "Something went wrong while setting key... " + err.Error(),
		})
	}

	newMessage, err := h.Service.SendNewMessage(messageRequest)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message": "Something went wrong while sending message... " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(newMessage)

}

func (h *MessageHandler) GetMessagesForUser(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	conversations, err := h.Service.GetMessagesForUser(userId)

	if err != nil {
		if err != nil {
			return ctx.Status(500).JSON(&fiber.Map{
				"message": "Something went wrong while getting messages.. " + err.Error(),
			})
		}

	}

	return ctx.Status(200).JSON(conversations)
}
