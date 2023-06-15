package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type UserHandler struct {
	Service        *services.UserStorage
	SessionService *services.SessionStorage
}

func NewUserHandler(service *services.UserStorage, sessionService *services.SessionStorage) *UserHandler {
	return &UserHandler{
		Service:        service,
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

	if err != nil {
		fmt.Println(err.Error())
		return c.Status(500).JSON(fiber.Map{
			"message": fmt.Sprintf("Error creating user: %s", err.Error()),
		})
	}

	go func() {
		fmt.Println(newUser.Id)
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

	return c.Status(fiber.StatusCreated).JSON(services.UserSessionResponse{
		Id:        newUser.Id,
		Username:  newUser.Username,
		Email:     newUser.Email,
		Image:     newUser.Image,
		Latitude:  newUser.Latitude,
		Longitude: newUser.Longitude,
		Session:   <-s1,
	})
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
	return c.Status(fiber.StatusCreated).JSON(newUser)
}

func (h *UserHandler) FetchUsersFeed(ctx *fiber.Ctx) error {

	userId := ctx.Params("userId")

	users, err := h.Service.GetAllUsers(userId)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": fmt.Sprintf("Error fetching users: %s", err.Error()),
		})
	}

	return ctx.Status(200).JSON(users)
}

// synchronous testing..

func (h *UserHandler) TestUserContextAggreationHandler(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")

	result, err := h.Service.SynchronousUserContextAggregation(userId)

	if err != nil {
		fmt.Println(err)
		return ctx.Status(500).JSON(&fiber.Map{
			"message": "Fucked up ",
		})
	}

	return ctx.Status(200).JSON(result)
}

// hi
