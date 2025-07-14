package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log/slog"
)

type Config struct {
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	Username string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"sla"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

type HTTPConfig struct {
	Address string `env:"HTTP_PORT" envDefault:"8080"`
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST" envDefault:"localhost"`
	Port string `env:"REDIS_PORT" envDefault:"6379"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type RabbitMQConfig struct {
	Host     string `env:"RABBITMQ_HOST" envDefault:"localhost"`
	Port     string `env:"RABBITMQ_PORT" envDefault:"5672"`
	Username string `env:"RABBITMQ_USER" envDefault:"guest"`
	Password string `env:"RABBITMQ_PASSWORD" envDefault:"guest"`
	Queue    string `env:"RABBITMQ_QUEUE" envDefault:"timestamp_events"`
}

func (c RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.Username, c.Password, c.Host, c.Port)
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found", "err", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
