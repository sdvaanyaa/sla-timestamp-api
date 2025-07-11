package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/sdvaanyaa/sla-timestamp-api/docs"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/config"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/handler"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository/postgres"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/pgdb"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("error", err))
		os.Exit(1)
	}

	client, err := pgdb.New(cfg.Postgres, log)
	if err != nil {
		log.Error("create postgres client failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer client.Close()

	storage := postgres.New(client)
	svc := service.New(storage)
	app := fiber.New()
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
