package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/microservices/shared/config"
	_ "github.com/microsoft/go-mssqldb"
)

type DB struct {
	*sql.DB
}

func Connect(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("sqlserver", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) ExecMigration(ctx context.Context, query string) error {
	_, err := db.ExecContext(ctx, query)
	return err
}
