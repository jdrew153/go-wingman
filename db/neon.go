package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func CreateNeonConnection(lc fx.Lifecycle) *pgxpool.Pool {
	err := godotenv.Load("../cmd/.env")
	conn, err := pgxpool.New(context.TODO(), os.Getenv("DATABASE_URL"))
	fmt.Println("DATABASE_URL", os.Getenv("DATABASE_URL"))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err != nil {
				return err
			}

			return nil
		},

		OnStop: func(ctx context.Context) error {
			conn.Close()
			return nil
		},
	})

	return conn
}
