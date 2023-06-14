package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type PostHandler struct {
	PostService *services.PostStorage
}

func NewPostHandler(p *services.PostStorage) *PostHandler {
	return &PostHandler{
		PostService: p,
	}
}

func (h *PostHandler) UploadNewPost(ctx *fiber.Ctx) error {
	var request services.NewPostRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	post, err := h.PostService.CreatePost(request)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(post)
}