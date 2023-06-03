package db

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/fx"
)


func CreateKafkaProducer(lc fx.Lifecycle) *kafka.Writer {
	err := godotenv.Load("../cmd/.env")

	mechanism, err := scram.Mechanism(scram.SHA256, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"))
	
	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS: &tls.Config{},
	}

	writer :=  kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("KAFKA_URL")},
		Topic: os.Getenv("KAFKA_TOPIC"),
		Balancer: &kafka.LeastBytes{},
		Dialer: dialer,
	})

	println("writer", writer)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err != nil {
				println("err", err)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			writer.Close()
			return nil
		},
	})

	return writer
}