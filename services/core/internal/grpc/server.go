package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func NewServer(port string) *Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	reflection.Register(s) // grpcurl 등 디버깅 도구 지원

	return &Server{
		grpcServer: s,
		port:       port,
	}
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", s.port, err)
	}
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
