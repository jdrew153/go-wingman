package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type CFImageUploaderHandler struct {
	Service *services.CFImageUploaderService
}


func NewCFImageUploaderHandler(service *services.CFImageUploaderService) *CFImageUploaderHandler {
	return &CFImageUploaderHandler{
		Service: service,
	}
}



func (h *CFImageUploaderHandler) CreateUploadUrl(c *fiber.Ctx) error {
	url, err := h.Service.CreateImageUploadURL()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return c.SendString(url)

}