package comptest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
)

type HealthCheck interface {
	Check(ctx context.Context) error
}

type HTTPHealthCheck struct {
	Addr string
}

func (c HTTPHealthCheck) Check(ctx context.Context) error {
	c.Addr = cleanAddress(c.Addr)

	req, err := http.NewRequest("GET", c.Addr, nil)
	if err != nil {
		return fmt.Errorf("failed to prepare the request: %w", err)
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to run the request: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

type TCPHealthCheck struct {
	Addr string
}

func (c TCPHealthCheck) Check(ctx context.Context) error {
	c.Addr = cleanAddress(c.Addr)

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", c.Addr)

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	conn.Close()
	return nil
}

func WaitForAll(ctx context.Context, checks ...HealthCheck) error {
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

func WaitForMainServer(ctx context.Context) error {
	mainHealthCheck := HTTPHealthCheck{
		Addr: fmt.Sprintf("http://localhost%s/readiness", os.Getenv("METRICS_ADDR")),
	}
	if err := WaitForAll(ctx, mainHealthCheck); err != nil {
		return err
	}
	return nil
}
