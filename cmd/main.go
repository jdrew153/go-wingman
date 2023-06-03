package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jdrew153/db"
	"github.com/jdrew153/handlers"
	"github.com/jdrew153/services"
	"go.uber.org/fx"
)


func newFiberServer(lc fx.Lifecycle, userHandler *handlers.UserHandler) *fiber.App {

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	group := app.Group("/api/v1/auth")
	group.Post("/signup", userHandler.CreateUser)



	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go app.Listen(":3001")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})

	return app
}

func main() {
	fx.New(
		fx.Provide(
			db.CreateNeonConnection,
			db.CreateRedisConnection,
			db.CreateKafkaProducer,
			services.NewUserService,
			handlers.NewUserHandler,
		),
		fx.Invoke(newFiberServer),
	).Run()
}