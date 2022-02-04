package mysql

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/ingridhq/comptest/db/dbutil"
	"github.com/jmoiron/sqlx"

	"github.com/ingridhq/comptest"
)

const schema = "mysql"

var _ comptest.HealthCheck = database{}

type database struct {
	dsn string
}

func Database(dsn string) *database {
	return &database{dsn: dsn}
}

func (c database) Check(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, schema, c.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	db.Close()
	return nil
}

func (c database) RunUpMigrations(migrationsSource string) error {
	return dbutil.RunUpMigrations(migrationsSource, fmt.Sprintf("%s://%s", schema, c.dsn))
}

func (c database) RunDownMigrations(migrationsSource string) error {
	return dbutil.RunDownMigrations(migrationsSource, fmt.Sprintf("%s://%s", schema, c.dsn))
}

// CreateDatabase will wait for db connection. You don't have to use WaitForAll() before CreateDatabase()
func (c database) CreateDatabase(ctx context.Context, dsn string) error {
	dbbase, dbname := dbutil.SplitDSN(dsn)

	if err := comptest.WaitForAll(ctx, c); err != nil {
		return err
	}

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
