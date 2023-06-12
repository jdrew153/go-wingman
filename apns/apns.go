package apns

import (
	"context"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"go.uber.org/fx"
)

func CreateAPNSService(lc fx.Lifecycle) *apns2.Client {
	authkey, err := token.AuthKeyFromFile("../AuthKey_ALS2696T7F.p8")

	if err != nil {
		panic(err)
	}

	token := &token.Token{
		AuthKey: authkey,
		KeyID:   "ALS2696T7F",
		TeamID:  "998F8YT622",
	}

	client := apns2.NewTokenClient(token)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err != nil {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return client
}