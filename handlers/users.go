package handlers

import (
	
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type UserHandler struct {
	Service *services.UserStorage
	SessionService *services.SessionStorage
}

func NewUserHandler(service *services.UserStorage, sessionService *services.SessionStorage) *UserHandler {
	return &UserHandler{
		Service: service,
		SessionService: sessionService,
	}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var user services.NewUser
	s1 := make(chan services.Session, 1)

	err := c.BodyParser(&user)

	if err != nil {
		return c.SendStatus(400)
	}

	newUser, err := h.Service.CreateNewUser(user)

	 go func() {
		newSession, err := h.SessionService.CreateSession(newUser.Id)

		if err != nil {
			println(err.Error())
			return 
		}

		s1 <- newSession
		
	}()

	if err != nil {
		return c.SendString(err.Error())
	}
	
	return c.JSON(services.UserSessionResponse{User: newUser, Session: <-s1})
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