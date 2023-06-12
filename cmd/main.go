package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jdrew153/apns"
	"github.com/jdrew153/db"
	"github.com/jdrew153/god"
	"github.com/jdrew153/handlers"
	"github.com/jdrew153/lib"
	"github.com/jdrew153/middleware"
	"github.com/jdrew153/services"
	"go.uber.org/fx"
)


func newFiberServer(
	lc fx.Lifecycle, 
	userHandler *handlers.UserHandler,
	interestHandler *handlers.InterestHandler,
	authHandler *handlers.AuthHandler,
	wingmanHandler *handlers.WingmanHandler,
	matchHandler *handlers.MatchHandler,
	notificationHandler *handlers.NotificationHandler,
	middleware *middleware.SessionMiddlewareHandler,
	) *fiber.App {

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	group := app.Group("/api/v1/auth")
	group.Post("/signup", userHandler.CreateUser)
	group.Post("/login", authHandler.AuthenticateUser)
	group.Post("/batch-create-interests", interestHandler.CreateBatchInterests)


	app.Post("/api/v1/wingman", wingmanHandler.CreateWingmanResponse)

	
	/// Notifications group
	notificationsGroup := app.Group("/api/v1/notifications")
	notificationsGroup.Post("/new", notificationHandler.CreateNotificationPair)
	notificationsGroup.Get("/user/:userId", notificationHandler.FetchDeviceTokenForUser)

	app.Use(middleware.AuthCheck)


	/// Authenticated routes
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "You are authenticated",
		})
	})


	/// Users group
	usersGropup := app.Group("/api/v1/users")
	usersGropup.Post("/feed", userHandler.FetchUsersFeed)

	//// Matches group
	matchesGroup := app.Group("/api/v1/matches")
	matchesGroup.Post("/new", matchHandler.CreateMatch)
	matchesGroup.Post("/update", matchHandler.FindAndUpDateMatchHandler)



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
			god.CreateGodClient,
			apns.CreateAPNSService,
			lib.NewMailer,
			services.NewUserService,
			services.NewSessionService,
			services.InterestService,
			services.NewAuthService,
			services.NewWingmanService,
			services.NewMatchService,
			services.NewNotificationService,
			handlers.NewUserHandler,
			handlers.NewAuthHandler,
			handlers.NewInterestHandler,
			handlers.NewWingmanHandler,
			handlers.NewMatchHandler,
			handlers.NewNotificationHandler,
			middleware.NewSessionMiddlewareHandler,
		),
		fx.Invoke(newFiberServer),
	).Run()
}