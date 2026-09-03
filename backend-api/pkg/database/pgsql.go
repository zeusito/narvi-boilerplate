package database

import (
	"backend-api/pkg/configurer"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun/extra/bunzerolog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type DatabaseConnection struct {
	Conn *bun.DB
	pool *pgxpool.Pool
}

func MustCreatePooledConnection(dbConfig configurer.DatabaseConfigurations) *DatabaseConnection {
	if !dbConfig.Enabled {
		log.Warn().Msg("database is disabled")
		return &DatabaseConnection{}
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName)

	parsedCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal().Msgf("Error parsing database configuration: %v", err)
		return nil
	}

	// Pool settings, min and max connections will be the same, effectively creating a fixed size pool
	parsedCfg.MaxConns = int32(dbConfig.PoolSize) //nolint:gosec
	parsedCfg.MinConns = int32(dbConfig.PoolSize) //nolint:gosec
	parsedCfg.MaxConnLifetime = 3 * time.Minute

	// Init a connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), parsedCfg)
	if err != nil {
		log.Fatal().Msgf("Error creating database connection pool: %v", err)
		return nil
	}

	// Init a connection compatible with the standard library
	sqlDB := stdlib.OpenDBFromPool(pool)

	log.Info().Msg("Successfully connected to database")

	db := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	hook := bunzerolog.NewQueryHook(
		bunzerolog.WithQueryLogLevel(zerolog.DebugLevel),
		bunzerolog.WithSlowQueryLogLevel(zerolog.WarnLevel),
		bunzerolog.WithErrorQueryLogLevel(zerolog.ErrorLevel),
		bunzerolog.WithSlowQueryThreshold(3*time.Second),
	)

	db = db.WithQueryHook(hook)

	return &DatabaseConnection{
		Conn: db,
		pool: pool,
	}
}

func (c *DatabaseConnection) Close() {
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	if c.pool != nil {
		c.pool.Close()
	}
}
