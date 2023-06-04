package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdrew153/services"
	"github.com/redis/go-redis/v9"
)

type SessionMiddlewareHandler struct {
	Con  *pgxpool.Pool
	RCon *redis.Client
}

func NewSessionMiddlewareHandler(con *pgxpool.Pool, rcon *redis.Client) *SessionMiddlewareHandler {
	return &SessionMiddlewareHandler{
		Con:  con,
		RCon: rcon,
	}
}


func (s *SessionMiddlewareHandler) AuthCheck(ctx *fiber.Ctx) error {
	var session services.Session

	header := ctx.Get("Authorization")

	if header != "" {
		fmt.Println(header)
		val, err := s.RCon.Get(ctx.Context(), header).Result()

		if err != nil {
			println(err.Error())
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
		println(val)
		json.Unmarshal([]byte(val), &session)
		now := time.Now().Unix()
		i, err := strconv.Atoi(session.Expires)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
		if now > int64(i) {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	} else {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	return ctx.Next()
}