package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
	"github.com/sashabaranov/go-openai"
)

type WingmanHandler struct {
	WingmanService *services.WingmanService
}

func NewWingmanHandler(w *services.WingmanService) *WingmanHandler {
	return &WingmanHandler{
		WingmanService: w,
	}
}

type WingmanRequest struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type WingmanRequestWithContext struct {
	WingmanRequest []openai.ChatCompletionMessage `json:"wingmanRequest"`
	Context        services.UserContext           `json:"context"`
}

func (h *WingmanHandler) CreateWingmanResponse(ctx *fiber.Ctx) error {

	var messagesWithContext WingmanRequestWithContext

	err := ctx.BodyParser(&messagesWithContext)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	return ctx.Status(201).JSON(h.WingmanService.CreateWingmanResponse(messagesWithContext.WingmanRequest,
		messagesWithContext.Context))
}
