package comptest

import (
	"fmt"
	"log"
	"net"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
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

func PubsubMessageToProtoMessage(m *pubsub.Message) (ptypes.DynamicAny, error) {
	a := &any.Any{}
	if err := proto.Unmarshal(m.Data, a); err != nil {
		return ptypes.DynamicAny{}, fmt.Errorf("could not unmarshall message: %w", err)
	}

	da := ptypes.DynamicAny{}
	if err := ptypes.UnmarshalAny(a, &da); err != nil {
		return ptypes.DynamicAny{}, fmt.Errorf("could not unmarshall any: %w", err)
	}

	return da, nil
}
