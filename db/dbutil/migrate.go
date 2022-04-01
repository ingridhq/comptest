package dbutil

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunUpMigrations will run all up migrations.
// * Assume database is ready for connections
// * Most common migration source: "file://migrations" (import "github.com/golang-migrate/migrate/v4/source/file")
// * dsn should start with database name: "mysql://root@tcp(localhost:3306)/db" (import "github.com/golang-migrate/migrate/v4/database/mysql")
func RunUpMigrations(migrationsSource, dsn string) error {
	m, err := migrate.New(migrationsSource, dsn)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil
}

func MustRunUpMigrations(migrationsSource, dsn string) {
	if err := RunUpMigrations(migrationsSource, dsn); err != nil {
		log.Fatalf("Error while doing migration: %v", err)
	}
	log.Println("Migrations were applied")
}

// RunDownMigrations will run all down migrations.
// * Assume database is ready for connections
// * Most common migration source: "file://migrations" (import "github.com/golang-migrate/migrate/v4/source/file")
// * dsn should start with database name: "mysql://root@tcp(localhost:3306)/db" (import "github.com/golang-migrate/migrate/v4/database/mysql")
func RunDownMigrations(migrationsSource, dsn string) error {
	m, err := migrate.New(migrationsSource, dsn)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil

}

func MustRunDownMigrations(migrationsSource, dsn string) {
	if err := RunDownMigrations(migrationsSource, dsn); err != nil {
		log.Fatalf("Error while doing migration: %v", err)
	}
	log.Println("Migrations were applied")
}
