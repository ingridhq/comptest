package comptest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ingridhq/comptest/binary"
	"golang.org/x/sync/errgroup"
)

type checker interface {
	Check(ctx context.Context) error
}

type comptest struct {
	m *testing.M

	timeout time.Duration

	checks     []checker
	beforeRun  []func() error
	buildPath  string
	binaryPath string
	logsPath   string
}

// New create new comptests suite.
func New(m *testing.M) *comptest {
	return &comptest{
		m:          m,
		binaryPath: os.TempDir() + "/main",
		logsPath:   "./comptest.logs",
		timeout:    30 * time.Second,
	}
}

// Build specifies main Go file to build.
func (c *comptest) Build(buildPath string) {
	c.buildPath = buildPath
}

// Wait adds check to perform before running tests.
func (c *comptest) Wait(checks ...checker) {
	c.checks = append(c.checks, checks...)
}

// BeforeRun sets action to perform before tests.
func (c *comptest) BeforeRun(fn func() error) {
	c.beforeRun = append(c.beforeRun, fn)
}

// Run builds, checks requirements and runs binary and tests.
func (c *comptest) Run(ctx context.Context) error {
	if c.buildPath == "" {
		return fmt.Errorf("call to Build is required")
	}
	if len(c.checks) < 1 {
		return fmt.Errorf("at least one check is required")
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := binary.BuildBinary(c.buildPath, c.binaryPath); err != nil {
		return err
	}

	cleaner, err := binary.RunBinary(c.binaryPath, c.logsPath)
	if err != nil {
		return err
	}
	defer cleaner()

	if err := waitForAll(ctx, c.checks...); err != nil {
		return err
	}

	for _, fn := range c.beforeRun {
		if err := fn(); err != nil {
			return err
		}
	}

	c.m.Run()

	return nil
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
