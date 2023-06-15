package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jdrew153/services"
)

type MatchHandler struct {
	MatchService          *services.MatchStorage
	NotificationService   *services.NotificationStorage
	WMNotificationService *services.WMNotificationStorage
}

func NewMatchHandler(m *services.MatchStorage, n *services.NotificationStorage, wm *services.WMNotificationStorage) *MatchHandler {
	return &MatchHandler{
		MatchService:          m,
		NotificationService:   n,
		WMNotificationService: wm,
	}
}

func (h *MatchHandler) CreateMatch(ctx *fiber.Ctx) error {
	var request services.MatchRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}
	fmt.Println(request)
	match, err := h.MatchService.CreateNewMatch(request)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(201).JSON(match)
}

type MatchUpdateRequest struct {
	
	UserIdA     string `json:"userIdA"`
	UserIdB     string `json:"userIdB"`
	MatchStatus string `json:"matchStatus"`
}

func (h *MatchHandler) FindAndUpDateMatchHandler(ctx *fiber.Ctx) error {
	var request MatchUpdateRequest

	err := ctx.BodyParser(&request)

	if err != nil {
		return ctx.Status(422).JSON(fiber.Map{
			"message": "Request body malformed",
		})
	}

	fmt.Println(request)

	match, err := h.MatchService.FindAndUpdateMatchStatus(request.UserIdA, request.UserIdB, request.MatchStatus)

	if match.MatchStatus == services.Mutual {

		errorChan := make(chan error, 1)

		go func() {

			go func() {
				result, err := h.WMNotificationService.CreateNotification(services.WMNotificationRequest{
					NotifiedUserId:   request.UserIdA,
					NotifierUserId:   request.UserIdB,
					NotificationType: services.MatchNotification,
					MatchId:          match.MatchId,
					AckStatus:        services.Unread,
				})

				if err != nil {
					errorChan <- err
				}
				fmt.Println(result)
			}()

			result, err := h.NotificationService.CreateAPNSNotification("You have a new match!", request.UserIdA)

			if err != nil {
				errorChan <- err
			}

			fmt.Println(result)
		}()

		go func() {
			go func() {
				result, err := h.WMNotificationService.CreateNotification(services.WMNotificationRequest{
					NotifiedUserId:   request.UserIdA,
					NotifierUserId:   request.UserIdB,
					NotificationType: services.MatchNotification,
					MatchId:          match.MatchId,
					AckStatus:        services.Unread,
				})

				if err != nil {
					errorChan <- err
				}
				fmt.Println(result)
			}()

			result, err := h.NotificationService.CreateAPNSNotification("You have a new match!", request.UserIdB)

			if err != nil {
				errorChan <- err
			}

			fmt.Println(result)
		}()

		err = <-errorChan

		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{
				"message": "Something went wrong " + err.Error(),
			})
		}
	}

	

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"message": "Something went wrong " + err.Error(),
		})
	}

	return ctx.Status(200).JSON(match)
}
