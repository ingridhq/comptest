package waitfor

import (
	"context"
	"fmt"
	"net/http"
)

func HTTP(addr string) httpHealthCheck {
	return httpHealthCheck{addr: addr}
}

type httpHealthCheck struct {
	addr string
}

func (c httpHealthCheck) String() string {
	return fmt.Sprintf("[HTTPCheck: %s]", c.addr)
}

func (c httpHealthCheck) Check(ctx context.Context) error {
	req, err := http.NewRequest("GET", c.addr, nil)
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
