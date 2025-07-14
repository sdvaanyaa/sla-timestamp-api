package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"
	_ "github.com/sdvaanyaa/sla-timestamp-api/docs"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/config"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/handler"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/middleware"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository/postgres"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/broker/rabbitmq"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache/rdscache"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/pgdb"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("error", err))
		os.Exit(1)
	}

	postgresClient, err := pgdb.New(cfg.Postgres, log)
	if err != nil {
		log.Error("create postgres client failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer postgresClient.Close()
	storage := postgres.New(postgresClient)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})
	cache, err := rdscache.New(redisClient, log)
	if err != nil {
		log.Error("create redis cache failed", slog.Any("error", err))
		os.Exit(1)
	}

	broker, err := rabbitmq.New(cfg.RabbitMQ.URL(), cfg.RabbitMQ.Queue, log) // New
	if err != nil {
		log.Error("create rabbitmq broker failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err = broker.Close(); err != nil {
			log.Error("close rabbitmq broker failed", slog.Any("error", err))
		}
	}()

	val := validator.New()
	svc := service.New(storage, val, cache, broker)

	app := fiber.New()
	app.Use(middleware.Logging(log))
	app.Use(middleware.RateLimiter())
	handler.New(app, svc)

	app.Get("/swagger/*", swagger.HandlerDefault)

	go func() {
		if err = app.Listen(":" + cfg.HTTP.Address); err != nil {
			log.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	if err = app.Shutdown(); err != nil {
		log.Error("shutdown failed", slog.Any("error", err))
	}
}
