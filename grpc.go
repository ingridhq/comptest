package comptest

import (
	"fmt"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
)

// MustStartGRPCServer will register and start grpc server.
func MustStartGRPCServer(addr string, regFn func(s *grpc.Server)) {
	addr = cleanAddress(addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	regFn(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

// CreateGRPCConn will create grpc conn with disabled TLS.
func CreateGRPCConn(addr string) (*grpc.ClientConn, error) {
	connOpts := getOptionsFromAddress(addr)
	addr = cleanAddress(addr)

	conn, err := grpc.Dial(addr, connOpts...)
	if err != nil {
		return nil, fmt.Errorf("create GRPC Conn error: %v", err)
	}
	return conn, nil
}

func MustCreateGRPCConn(addr string) *grpc.ClientConn {
	conn, err := CreateGRPCConn(addr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

// cleanAddress will return address without options
func cleanAddress(addr string) string {
	return strings.Split(addr, "?")[0]
}

func getOptionsFromAddress(addr string) []grpc.DialOption {
	connOpts := []grpc.DialOption{}

	if strings.Contains(addr, "insecure=true") {
		connOpts = append(connOpts, grpc.WithInsecure())
	}

	return connOpts
}
