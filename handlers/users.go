package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type UserHandler struct {
	Service *services.UserStorage
}

func NewUserHandler(service *services.UserStorage) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var user services.NewUser

	err := c.BodyParser(&user)

	if err != nil {
		return c.SendStatus(400)
	}

	newUser, err := h.Service.CreateNewUser(user)

	if err != nil {
		return c.SendString(err.Error())
	}
	return c.JSON(newUser)
}


func (h *UserHandler) CacheUser(c *fiber.Ctx) error {
	var user services.NewUser

	err := c.BodyParser(&user)

	if err != nil {
		return c.SendStatus(400)
	}

	newUser, err := h.Service.CacheUser(user)

	if err != nil {
		return c.SendString(err.Error())
	}
	return c.JSON(newUser)
}