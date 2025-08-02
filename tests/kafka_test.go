package tests

import (
	"URL-Shortener/internal/config"
	"URL-Shortener/internal/event"
	"URL-Shortener/internal/kafka"
	"log/slog"
	"os"
	"testing"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func TestKafkaSendMessage(t *testing.T) {
	if os.Getenv("KAFKA_TEST") != "1" {
		t.Skip("skipping Kafka integration test (KAFKA_TEST != 1)")
	}

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	p, err := kafka.NewProducer(cfg.KafkaProducerAdapter, log)
	if err != nil {
		t.Fatalf("failed to create kafka producer: %v", err)
	}
	defer p.CloseProducer()

	test_event := event.AuthEvent{
		Type: "login",
		Email: "testemail@gmail.com",
		TimeStamp: time.Now(),
	}

	err = p.ProduceEvent(test_event, "auth-topic")
	if err != nil {
		t.Fatalf("failed to produce event: %v", err)
	}

	time.Sleep(1 * time.Second)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}