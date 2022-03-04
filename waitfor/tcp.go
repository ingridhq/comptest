package waitfor

import (
	"context"
	"fmt"
	"net"
)

func TCP(addr string) tcpHealthCheck {
	return tcpHealthCheck{
		addr: addr,
	}
}

type tcpHealthCheck struct {
	addr string
}

func (c tcpHealthCheck) Check(ctx context.Context) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", c.addr)

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	conn.Close()
	return nil
}
