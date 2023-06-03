package db

import (
	"os"
	"context"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func CreateRedisConnection(lc fx.Lifecycle) *redis.Client {

	err := godotenv.Load("../cmd/.env")
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))

	client := redis.NewClient(opt)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err != nil {
				println("err", err)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			client.Close()
			return nil
		},
	})
	
	return client
}