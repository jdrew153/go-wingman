package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type WMMessagingHandler struct {
	Service *services.WMMessagingService
}

func NewWMMessagingHandler(s *services.WMMessagingService) *WMMessagingHandler {

	return &WMMessagingHandler{
		s,
	}
}

func (h *WMMessagingHandler) CreateNewMessageWithContext(ctx *fiber.Ctx) error {
	var message services.WMMessageRequest

	err := ctx.BodyParser(&message)

	if err != nil {
		return ctx.Status(422).JSON(&fiber.Map{
			"message": "Malformed request body " + err.Error(),
		})
	}

	fmt.Println(message.ReceiverId)

	if len(message.ConversationId) < 2 {
		conversationId, err := h.Service.SearchAndUpsertANewConversation(message.SenderId, message.ReceiverId)
		if err != nil {
			return ctx.Status(500).JSON(&fiber.Map{
				"message": "Something went wrong... " + err.Error(),
			})
		}
		message.ConversationId = conversationId
	}

	result, err := h.Service.AddNewMessageToConversation(message)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message": "Something went wrong... " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(result)
}

func (h *WMMessagingHandler) GetConversationsForUser(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	results, err := h.Service.FetchConversationsForUser(userId)

	if err != nil {
		return ctx.Status(500).JSON(&fiber.Map{
			"message": "Something went wrong... " + err.Error(),
		})
	}

	return ctx.JSON(results)
}
