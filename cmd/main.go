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
	postHandler *handlers.PostHandler,
	asyncHandler *handlers.AsyncHandler,
	messageHandler *handlers.MessageHandler,
	wmmessageHandler *handlers.WMMessagingHandler,
	cfHandler *handlers.CFImageUploaderHandler,
	middleware *middleware.SessionMiddlewareHandler,
) *fiber.App {

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	group := app.Group("/api/v1/auth")

	group.Post("/login", authHandler.AuthenticateUser)
	group.Post("/batch-create-interests/:userId", interestHandler.CreateBatchInterests)
	app.Post("/api/v1/register", userHandler.CreateUser)

	app.Post("/api/v1/wingman", wingmanHandler.CreateWingmanResponse)

	app.Get("api/v1/image/upload", cfHandler.CreateUploadUrl)

	// sync testing
	app.Get("/api/v1/sync/users/:userId", userHandler.TestUserContextAggreationHandler)

	// async testing
	app.Get("/api/v1/async/users/:userId", asyncHandler.TestableAsyncFunction)

	// Posts Group
	postsGroup := app.Group("/api/v1/posts")
	postsGroup.Post("/new-post", postHandler.UploadNewPost)

	/// Notifications group
	notificationsGroup := app.Group("/api/v1/notifications")
	notificationsGroup.Post("/new", notificationHandler.CreateNotificationPair)
	notificationsGroup.Get("/user/:userId", notificationHandler.FetchDeviceTokenForUser)
	notificationsGroup.Post("/send/:userId", notificationHandler.SendAPNSNotification)

	// Messaging Group
	messageGroup := app.Group("/api/v1/messages")
	messageGroup.Post("/new-message", messageHandler.SendMessageComplete)
	messageGroup.Get("/:userId", messageHandler.GetMessagesForUser)

	// Testing sql backed messaging
	wmMessageGroup := app.Group("/api/v1/wmMessages")
	wmMessageGroup.Post("/new", wmmessageHandler.CreateNewMessageWithContext)
	wmMessageGroup.Get("/conversations/:userId", wmmessageHandler.GetConversationsForUser)

	app.Use(middleware.AuthCheck)

	/// Authenticated routes
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "You are authenticated",
		})
	})

	/// Users group
	usersGroup := app.Group("/api/v1/users")
	usersGroup.Post("/feed/:userId", userHandler.FetchUsersFeed)

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

//Hi

func main() {
	fx.New(
		fx.Provide(
			db.CreateNeonConnection,
			db.CreateRedisConnection,
			db.CreateKafkaProducer,
			god.CreateGodClient,
			apns.CreateAPNSService,
			lib.NewMailer,
			lib.CreateHttpClient,
			services.NewUserService,
			services.NewSessionService,
			services.InterestService,
			services.NewAuthService,
			services.NewWingmanService,
			services.NewMatchService,
			services.NewNotificationService,
			services.NewWMNotificationStorage,
			services.NewCFImageUploaderService,
			services.NewPostService,
			services.NewAsyncService,
			services.NewMessageService,
			services.NewWMMessagingServices,
			handlers.NewUserHandler,
			handlers.NewAuthHandler,
			handlers.NewInterestHandler,
			handlers.NewWingmanHandler,
			handlers.NewMatchHandler,
			handlers.NewNotificationHandler,
			handlers.NewWMNotificationHandler,
			handlers.NewCFImageUploaderHandler,
			handlers.NewPostHandler,
			handlers.NewAsyncHandler,
			handlers.NewMessageHandler,
			handlers.NewWMMessagingHandler,
			middleware.NewSessionMiddlewareHandler,
		),
		fx.Invoke(newFiberServer),
	).Run()
}
