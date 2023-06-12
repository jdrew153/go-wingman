package god

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/fx"
)

func CreateGodClient(lc fx.Lifecycle) *openai.Client {
	err := godotenv.Load("../cmd/.env")
	c := openai.NewClient(os.Getenv("OPENAI_KEY"))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err != nil {
				return err
			}

			return nil
		}, 
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return c
}