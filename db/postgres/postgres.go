package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ingridhq/comptest/db/dbutil"
	"github.com/jmoiron/sqlx"
)

const schema = "postgres"

type database struct {
	dsn string
}

// Database create Postgresql suite for database initialization.
func Database(dsn string) *database {
	return &database{dsn: dsn}
}

// Check implements checker interface for convenient use in HealthChecks function.
func (c database) Check(ctx context.Context) error {
	conninfo, _ := dbutil.SplitDSN(c.dsn)
	conninfo = fmt.Sprintf("%s/postgres?sslmode=disable", conninfo)

	db, err := sqlx.ConnectContext(ctx, schema, prepareDSN(conninfo))
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	db.Close()
	return nil
}

// RunUpMigrations runs UP migrations from source.
func (c database) RunUpMigrations(migrationsSource string) error {
	return dbutil.RunUpMigrations(migrationsSource, prepareDSN(c.dsn))
}

// RunDownMigrations runs DOWN migrations from source.
func (c database) RunDownMigrations(migrationsSource string) error {
	return dbutil.RunDownMigrations(migrationsSource, prepareDSN(c.dsn))
}

func prepareDSN(dsn string) string {
	if strings.HasPrefix(dsn, fmt.Sprintf("%s://", schema)) {
		return dsn
	}
	return fmt.Sprintf("%s://%s", schema, dsn)
}

// CreateDatabase creates database extracted from DSN.
func (c database) CreateDatabase(ctx context.Context) error {
	conninfo, dbname := dbutil.SplitDSN(c.dsn)
	conninfo = fmt.Sprintf("%s/postgres?sslmode=disable", conninfo)

	// Connect to postgres database, to create new db. Sqlx require connection to database.
	db, err := sqlx.ConnectContext(ctx, schema, conninfo)
	if err != nil {
		return err
	}
	defer db.Close()

	type dbExist struct {
		Datname string `db:"datname"`
	}
	err = db.GetContext(ctx, &dbExist{}, fmt.Sprintf("SELECT datname FROM pg_catalog.pg_database WHERE datname='%v';", dbname))

	// Database exists when no errors.
	if err == nil {
		return nil
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// ErrNoRows means database do not exists
		default:
			return err
		}
	}

	// In postgres there is no CREATE DATABASE IF NOT EXISTS query
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %v;", dbname))
	if err != nil {
		return err
	}

	return nil
}
