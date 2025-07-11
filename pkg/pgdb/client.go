package pgdb

import (
	"context"
	"fmt"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/config"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Client struct {
	conn *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg config.PostgresConfig, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)

	log.Info("connecting to database",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("database", cfg.Database),
	)

	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error("error connecting to database", slog.Any("error", err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = conn.Ping(context.Background()); err != nil {
		log.Error("error pinging database", slog.Any("error", err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("database connection established")

	return &Client{
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
	c.log.Info("database connection closed")
}

func (c *Client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := c.conn.Query(ctx, sql, args...)

	c.logQuery(sql, time.Since(start), err)

	return rows, err
}

func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.conn.QueryRow(ctx, sql, args...)
}

func (c *Client) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := c.conn.Exec(ctx, sql, arguments...)

	c.logQuery(sql, time.Since(start), err)

	return tag, err
}

func (c *Client) logQuery(sql string, duration time.Duration, err error) {
	fields := strings.Fields(sql)
	operation := "UNKNOWN"
	if len(fields) > 0 {
		operation = strings.ToUpper(fields[0])
	}

	cleanSQL := strings.Join(fields, " ")

	attrs := []slog.Attr{
		slog.String("operation", operation),
		slog.String("sql", cleanSQL),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.LogAttrs(context.Background(), slog.LevelError, "query failed", attrs...)
		return
	}

	c.log.LogAttrs(context.Background(), slog.LevelDebug, "query succeeded", attrs...)
}
