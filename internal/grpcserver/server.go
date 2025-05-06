package grpcserver

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	grpcServer *grpc.Server
	health     *health.Server
	addr       string
}

func NewServer(addr string, opts ...grpc.ServerOption) *Server {
	s := grpc.NewServer(opts...)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)

	return &Server{
		grpcServer: s,
		health:     healthServer,
		addr:       addr,
	}
}

func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.grpcServer.RegisterService(desc, impl)
}

func (s *Server) Start() error {
	s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	slog.Info("Starting server on", "addr", s.addr)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.GracefulStop()
}

func (s *Server) ForceStop() {
	s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.Stop()
}
