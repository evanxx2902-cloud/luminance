package server

import (
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer creates a new gRPC server (scaffold only — no services registered yet).
func NewGRPCServer() *kratosgrpc.Server {
	return kratosgrpc.NewServer(
		kratosgrpc.Address(":9000"),
	)
}
