package database

import (
	"context"
	"fmt"
	"time"

	"github.com/inst-dev/webhook/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Pool wraps pgxpool.Pool for database operations
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Int32("max_conns", cfg.MaxConns).
		Msg("Database pool created")

	return &Pool{pool}, nil
}

// Close closes the database pool
func (p *Pool) Close() {
	p.Pool.Close()
}
