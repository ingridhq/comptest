package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"

	"github.com/ingridhq/comptest"
	"github.com/ingridhq/comptest/db/dbutil"
)

const schema = "postgres"

var _ comptest.HealthCheck = database{}

type database struct {
	dsn string
}

func Database(dsn string) *database {
	return &database{dsn: dsn}
}

func (c database) Check(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, schema, prepareDSN(c.dsn))
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	db.Close()
	return nil
}

func (c database) RunUpMigrations(migrationsSource string) error {
	return dbutil.RunUpMigrations(migrationsSource, prepareDSN(c.dsn))
}

func (c database) RunDownMigrations(migrationsSource string) error {
	return dbutil.RunDownMigrations(migrationsSource, prepareDSN(c.dsn))
}

func prepareDSN(dsn string) string {
	if strings.HasPrefix(dsn, fmt.Sprintf("%s://", schema)) {
		return dsn
	}
	return fmt.Sprintf("%s://%s", schema, dsn)
}

// CreateDatabase will wait for db connection. You don't have to use WaitForAll() before CreateDatabase()
func (c database) CreateDatabase(ctx context.Context) error {
	conninfo, dbname := dbutil.SplitDSN(c.dsn)
	conninfo = fmt.Sprintf("%s/postgres?sslmode=disable", conninfo)

	// Wait for postgres database
	postgresDB := Database(conninfo)
	if err := comptest.WaitForAll(ctx, postgresDB); err != nil {
		return err
	}

	// Connect to postgres database, to create new db. Sqlx require connection to database.
	db, err := sqlx.ConnectContext(ctx, schema, conninfo)
	if err != nil {
		return err
	}

	type dbExist struct {
		Datname string `db:"datname"`
	}
	err = db.GetContext(ctx, &dbExist{}, fmt.Sprintf("SELECT datname FROM pg_catalog.pg_database WHERE datname='%v';", dbname))

	// TODO refactor that
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

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %v;", dbname))
	if err != nil {
		return err
	}

	return nil
}
