package comptest

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ingridhq/comptest/binary"
	"golang.org/x/sync/errgroup"
)

type checker interface {
	Check(ctx context.Context) error
}

type CleanupFunc func()

type comptest struct {
	timeout time.Duration

	binaryPath string
	logsPath   string
}

// New create new comptests suite.
func New() *comptest {
	return &comptest{
		binaryPath: os.TempDir() + "/main",
		logsPath:   "./comptest.logs",
		timeout:    30 * time.Second,
	}
}

// SetBinaryPath sets custom path where binary will be build.
func (c *comptest) SetBinaryPath(binaryPath string) {
	c.binaryPath = binaryPath
}

// SetLogsPath sets custom file to store logs from binary.
func (c *comptest) SetLogsPath(logsPath string) {
	c.logsPath = logsPath
}

// HealthChecks waits for external dependencies (PubSubs, Databases, GRPC mocks) to be ready.
func (c *comptest) HealthChecks(checks ...checker) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	if err := waitForAll(ctx, checks...); err != nil {
		log.Fatalf("Failed to check external dependencies: %v", err)
	}
}

// BuildAndRun builds, runs binary, waits for readiness check and runs tests.
// Returns cleanup function that needs to be invoked after tests.
func (c *comptest) BuildAndRun(buildPath string, readiness checker) CleanupFunc {
	if err := binary.BuildBinary(buildPath, c.binaryPath); err != nil {
		log.Fatalf("Failed to build binary: %v", err)
	}

	cleaner, err := binary.RunBinary(c.binaryPath, c.logsPath)
	if err != nil {
		log.Fatalf("Failed to run binary: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	if err := waitForAll(ctx, readiness); err != nil {
		cleaner()
		log.Fatalf("Failed to check readiness: %v", err)
	}

	return cleaner
}

func waitForAll(ctx context.Context, checks ...checker) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, c := range checks {
		c := c
		g.Go(func() error {
			return backoff.Retry(func() error {
				return c.Check(ctx)
			}, newExponentialBackOff(ctx))
		})
	}
	return g.Wait()
}

func newExponentialBackOff(ctx context.Context) backoff.BackOffContext {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     10 * time.Millisecond,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Second,
		MaxElapsedTime:      backoff.DefaultMaxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return backoff.WithContext(b, ctx)
}
