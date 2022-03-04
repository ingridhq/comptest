package waitfor

import (
	"context"
	"fmt"
	"net"
	"strings"
)

func TCP(addr string) tcpHealthCheck {
	return tcpHealthCheck{
		addr: addr,
	}
}

type tcpHealthCheck struct {
	addr string
}

func (c tcpHealthCheck) String() string {
	return fmt.Sprintf("[TCPCheck: %s]", c.addr)
}

func (c tcpHealthCheck) Check(ctx context.Context) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", cleanupAddress(c.addr))

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	conn.Close()
	return nil
}

func cleanupAddress(addr string) string {
	return strings.Split(addr, "?")[0]
}
