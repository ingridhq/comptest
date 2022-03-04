package comptest

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cenkalti/backoff/v4"
	"github.com/ingridhq/comptest/binary"
	"golang.org/x/sync/errgroup"
)

type checker interface {
	Check(ctx context.Context) error
}

type CleanupFunc func()

type comptest struct {
	ctx context.Context

	binaryPath string
	logsPath   string
}

// New create new comptests suite.
func New(ctx context.Context) *comptest {
	return &comptest{
		binaryPath: os.TempDir() + "/main",
		logsPath:   "./comptest.log",
		ctx:        ctx,
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
	if err := waitForAll(c.ctx, checks...); err != nil {
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

	if err := waitForAll(c.ctx, readiness); err != nil {
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
				err := c.Check(ctx)

				select {
				case <-ctx.Done():
					if err != nil {
						return backoff.Permanent(fmt.Errorf("check %v failed: %w", c, err))
					}
					return nil
				default:
					if err != nil {
						return fmt.Errorf("check %v failed: %w", c, err)
					}
					return nil
				}
			}, backoff.NewExponentialBackOff())
		})
	}

	return g.Wait()
}
