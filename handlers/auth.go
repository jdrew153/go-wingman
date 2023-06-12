package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type AuthHandler struct {
	AuthSesion *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthSesion: authService,
	}
}


func (h *AuthHandler) AuthenticateUser(ctx *fiber.Ctx) error {
	var body services.AuthenticationModel

	err := ctx.BodyParser(&body)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	val, err := h.AuthSesion.AuthenticateUser(body)

	if err != nil {
		return ctx.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(200).JSON(val)
}