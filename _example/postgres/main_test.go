package comptest

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/ingridhq/comptest"
	"github.com/ingridhq/comptest/db/postgres"
)

type config struct {
	DBPostgresDSN string `envconfig:"DB_POSTGRES_DSN"`
}

func TestMain(t *testing.M) {
	if os.Getenv("RUN_COMPONENT_TESTS") != "true" {
		return
	}

	var cfg config
	envconfig.MustProcess("", &cfg)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	postgresDB := postgres.Database(cfg.DBPostgresDSN)
	if err := postgresDB.CreateDatabase(ctx); err != nil {
		log.Fatalf("could not create database: %v", err)
	}

	if err := comptest.WaitForAll(ctx,
		postgresDB,
	); err != nil {
		log.Fatalf("wait for all failed: %v", err)
	}

	t.Run()
}
