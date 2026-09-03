package testbox

import (
	"context"
	"database/sql"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	// Import the pgx stdlib compatibility layer
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	DefaultDBName     = "testdb"
	DefaultDBUserName = "testuser"
	DefaultDBPass     = "qwerty1234"
)

type ConnectionData struct {
	DBName   string
	UserName string
	Password string // nolint:gosec
	Host     string
	Port     int
}

func InitPostgresqlContainer(ctx context.Context, initScripts []string) (*bun.DB, func(), error) {
	// Initialize PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase(DefaultDBName),
		postgres.WithUsername(DefaultDBUserName),
		postgres.WithPassword(DefaultDBPass),
		postgres.WithOrderedInitScripts(initScripts...),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		return nil, nil, err
	}

	// Extract connection properties and connect
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	dbClient, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() {
		_ = dbClient.Close()
		_ = testcontainers.TerminateContainer(pgContainer)
	}

	return bun.NewDB(dbClient, pgdialect.New(), bun.WithDiscardUnknownColumns()), closeFunc, nil
}
