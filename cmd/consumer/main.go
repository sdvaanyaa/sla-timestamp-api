package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/config"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/broker/rabbitmq"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache/rdscache"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := loadConfig(log)
	cache := initCache(cfg, log)
	broker := initBroker(cfg, log)

	ch, q := initChannelAndQueue(broker, cfg, log)
	msgs := initConsume(ch, q, log)

	go consumeMessages(cache, msgs, log)

	waitForSignal(log, broker, ch)
}

func loadConfig(log *slog.Logger) *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("config load failed", slog.Any("error", err))
		os.Exit(1)
	}
	return cfg
}

func initCache(cfg *config.Config, log *slog.Logger) cache.Cache {
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})
	cache, err := rdscache.New(redisClient, log)
	if err != nil {
		log.Error("create redis cache failed", slog.Any("error", err))
		os.Exit(1)
	}
	return cache
}

func initBroker(cfg *config.Config, log *slog.Logger) *rabbitmq.Client {
	broker, err := rabbitmq.New(cfg.RabbitMQ.URL(), cfg.RabbitMQ.Queue, log)
	if err != nil {
		log.Error("create rabbitmq broker failed", slog.Any("error", err))
		os.Exit(1)
	}
	return broker
}

func initChannelAndQueue(
	broker *rabbitmq.Client,
	cfg *config.Config,
	log *slog.Logger,
) (*amqp091.Channel, amqp091.Queue) {
	ch, err := broker.Channel()
	if err != nil {
		log.Error("channel failed", slog.Any("error", err))
		os.Exit(1)
	}

	q, err := ch.QueueDeclare(cfg.RabbitMQ.Queue, true, false, false, false, nil)
	if err != nil {
		log.Error("queue declare failed", slog.Any("error", err))
		os.Exit(1)
	}
	return ch, q
}

func initConsume(ch *amqp091.Channel, q amqp091.Queue, log *slog.Logger) <-chan amqp091.Delivery {
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Error("consume failed", slog.Any("error", err))
		os.Exit(1)
	}
	return msgs
}

func consumeMessages(cache cache.Cache, msgs <-chan amqp091.Delivery, log *slog.Logger) {
	for d := range msgs {
		var event map[string]any
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Error("unmarshal failed", slog.Any("error", err))
			continue
		}

		action, ok := event["action"].(string)
		if !ok {
			continue
		}

		switch action {
		case "create":
			handleCreate(event, cache, log)
		case "delete":
			handleDelete(event, cache, log)
		}
	}
}

func handleCreate(event map[string]any, cache cache.Cache, log *slog.Logger) {
	dataJSON, err := json.Marshal(event["data"])
	if err != nil {
		log.Error("marshal data failed", slog.Any("error", err))
		return
	}

	var ts entity.Timestamp
	if err := json.Unmarshal(dataJSON, &ts); err != nil {
		log.Error("unmarshal timestamp failed", slog.Any("error", err))
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf(service.TimestampCachePrefix, ts.ID.String())
	_ = cache.Set(ctx, key, &ts, service.CacheTTL)
}

func handleDelete(event map[string]any, cache cache.Cache, log *slog.Logger) {
	idStr, ok := event["id"].(string)
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("parse id failed", slog.Any("error", err))
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf(service.TimestampCachePrefix, id.String())
	_ = cache.Delete(ctx, key)
	_ = cache.Delete(ctx, service.ListCachePrefix)
}

func waitForSignal(log *slog.Logger, broker *rabbitmq.Client, ch *amqp091.Channel) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	if err := ch.Close(); err != nil {
		log.Error("channel close failed", slog.Any("error", err))
	}
	if err := broker.Close(); err != nil {
		log.Error("close rabbitmq broker failed", slog.Any("error", err))
	}

	log.Info("shutdown")
}
