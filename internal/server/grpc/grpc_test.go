package grpc

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/protos/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"sync"
	"testing"
	"time"
)

type mockMetricServiceServer struct {
	mock.Mock
	gen.UnimplementedMetricServiceServer
}

func (m *mockMetricServiceServer) GetAll(ctx context.Context, req *emptypb.Empty) (*gen.GetAllResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*gen.GetAllResponse), args.Error(1)
}

func (m *mockMetricServiceServer) SaveAll(ctx context.Context, req *gen.SaveAllRequest) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *mockMetricServiceServer) Find(ctx context.Context, req *gen.Metric) (*gen.Metric, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*gen.Metric), args.Error(1)
}

func (m *mockMetricServiceServer) Save(ctx context.Context, req *gen.Metric) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

type mockListener struct {
	mock.Mock
}

func (m *mockListener) Accept() (net.Conn, error) {
	args := m.Called()
	return args.Get(0).(net.Conn), args.Error(1)
}

func (m *mockListener) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockListener) Addr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func TestNewGRPCServer(t *testing.T) {
	addrConfig := config.AddrConfig{Host: "localhost", Port: 50051}
	logger := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
	server := NewGRPCServer(addrConfig, logger, "192.168.1.0/24")
	assert.NotNil(t, server)
	assert.Equal(t, "localhost:50051", server.addr)
	assert.NotNil(t, server.srv)
}

func TestRegister(t *testing.T) {
	addrConfig := config.AddrConfig{Host: "localhost", Port: 50051}
	logger := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
	server := NewGRPCServer(addrConfig, logger, "")

	var registeredServer gen.MetricServiceServer = &mockMetricServiceServer{}

	svc := &mockMetricServiceServer{}
	server.Register(svc)
	assert.Equal(t, svc, registeredServer)
}

func TestRunSuccess(t *testing.T) {
	server := &Server{
		addr: "localhost:50051",
		srv:  grpc.NewServer(),
	}
	defer server.srv.Stop()
	listener := &mockListener{}
	listener.On("Addr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50051})
	listener.On("Close").Return(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Run()
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	listener.On("Close").Return(nil).Maybe()
	server.Shutdown()
	wg.Wait()
}
