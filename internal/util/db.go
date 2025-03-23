// internal/util/db.go
package util

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jskallebak/prod/internal/db/sqlc"
)

// InitDB initializes a connection to the database
func InitDB() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	return pgxpool.New(context.Background(), dbURL)
}

// The caller is responsible for closing the returned dbpool.
func InitDBAndQueries(ctx context.Context) (*pgxpool.Pool, *sqlc.Queries, error) {
	dbpool, err := InitDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	queries := sqlc.New(dbpool)
	return dbpool, queries, nil
}

// InitDBAndQueriesCLI initializes a database connection for CLI commands,
// The caller is responsible for closing the returned dbpool.
func InitDBAndQueriesCLI() (*pgxpool.Pool, *sqlc.Queries, bool) {
	dbpool, err := InitDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		return nil, nil, false
	}

	queries := sqlc.New(dbpool)
	return dbpool, queries, true
}
