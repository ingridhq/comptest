package mysql

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/ingridhq/comptest/db/dbutil"
	"github.com/jmoiron/sqlx"
)

const schema = "mysql"

type database struct {
	dsn string
}

// Database create MySQL suite for database initialization.
func Database(dsn string) *database {
	return &database{dsn: dsn}
}

func (c database) String() string {
	return fmt.Sprintf("[MySQL: %s]", c.dsn)
}

// Check implements checker interface for convenient use in HealthChecks function.
func (c database) Check(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, schema, c.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	db.Close()
	return nil
}

// RunUpMigrations runs UP migrations from source.
func (c database) RunUpMigrations(migrationsSource string) error {
	return dbutil.RunUpMigrations(migrationsSource, fmt.Sprintf("%s://%s", schema, c.dsn))
}

// RunDownMigrations runs DOWN migrations from source.
func (c database) RunDownMigrations(migrationsSource string) error {
	return dbutil.RunDownMigrations(migrationsSource, fmt.Sprintf("%s://%s", schema, c.dsn))
}

// CreateDatabase creates database extracted from DSN.
func (c database) CreateDatabase(ctx context.Context, dsn string) error {
	dbbase, dbname := dbutil.SplitDSN(dsn)

	db, err := sqlx.ConnectContext(ctx, "mysql", dbbase+"/")
	if err != nil {
		return fmt.Errorf("failed to connect to database %q : %w", c.dsn, err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS ?", dbname)
	if err != nil {
		return fmt.Errorf("failed to create database %q: %w", dbname, err)
	}

	return nil
}
