package comptest

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/option"
)

type config struct {
	StorageHost   string `envconfig:"STORAGE_EMULATOR_HOST"`
	StorageBucket string `envconfig:"STORAGE_BUCKET"`
}

type Environment struct {
	Storage *storage.BucketHandle
}

var env Environment

func TestMain(t *testing.M) {
	if os.Getenv("RUN_COMPONENT_TESTS") != "true" {
		return
	}

	var cfg config
	envconfig.MustProcess("", &cfg)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gcs, err := storage.NewClient(ctx, option.WithEndpoint(fmt.Sprintf("http://%s", cfg.StorageHost)))
	if err != nil {
		log.Fatalf("failed to create datastore client: %v", err)
	}

	env = Environment{
		Storage: gcs.Bucket(cfg.StorageBucket),
	}

	t.Run()
}

func TestStorage(t *testing.T) {
	const (
		content = "blob"
		name    = "name"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mustStoreBlob(t, ctx, env.Storage, name, content)
	read := string(mustReadBlob(t, ctx, env.Storage, name))
	if content != read {
		t.Errorf("Expected read content to be: %q, got %q", content, read)
	}

	t.Log("Storage is ok")
}

func mustStoreBlob(t *testing.T, ctx context.Context, bkt *storage.BucketHandle, name, content string) {
	w := bkt.Object(name).NewWriter(ctx)

	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write blob: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close blob: %v", err)
	}
}

func mustReadBlob(t *testing.T, ctx context.Context, bkt *storage.BucketHandle, name string) []byte {
	r, err := bkt.Object(name).NewReader(ctx)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	bb, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read blob: %v", err)
	}

	return bb
}
