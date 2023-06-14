package lib

import (
	"context"
	"net/http"

	"go.uber.org/fx"
)

func CreateHttpClient(lc fx.Lifecycle) *http.Client {
	client := &http.Client{}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return client
}