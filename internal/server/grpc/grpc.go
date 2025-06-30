package grpc

import (
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/protos/gen"
	"github.com/MxTrap/metrics/internal/server/grpc/interceptor"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	srv  *grpc.Server
	addr string
}

func NewGRPCServer(addr config.AddrConfig, logger grpc.UnaryServerInterceptor, cidr string) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger,
			interceptors.StatusErrorInterceptor,
			interceptors.IpValidator(cidr),
		),
	)

	return &Server{
		srv:  srv,
		addr: fmt.Sprintf("%s:%d", addr.Host, addr.Port),
	}

}

func (s *Server) Register(svc gen.MetricServiceServer) {
	gen.RegisterMetricServiceServer(s.srv, svc)
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	err = s.srv.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown() {
	s.srv.GracefulStop()
}
