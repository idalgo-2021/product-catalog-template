package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds configurable parameters for the PostgreSQL connection pool.
type PoolConfig struct {
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

// DefaultPoolConfig returns a PoolConfig with sensible defaults.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:        10,
		MinConns:        2,
		MaxConnIdleTime: 30 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
	}
}

// NewPostgresPool creates a new connection pool for PostgreSQL with the given configuration
// and verifies the connection.
func NewPostgresPool(ctx context.Context, connectionURL string, cfg PoolConfig) (*pgxpool.Pool, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	dbpool, err := pgxpool.NewWithConfig(timeoutCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure database pool: %w", err)
	}

	if err := dbpool.Ping(timeoutCtx); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return dbpool, nil
}
