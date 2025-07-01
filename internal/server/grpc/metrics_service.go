package grpc

import (
	"context"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/protos/gen"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type saver interface {
	Save(ctx context.Context, metrics commonmodels.Metric) error
	SaveAll(ctx context.Context, metrics []commonmodels.Metric) error
}

type getter interface {
	Find(ctx context.Context, metric commonmodels.Metric) (commonmodels.Metric, error)
	GetAll(ctx context.Context) (map[string]any, error)
}

// MetricService определяет интерфейс для операций с метриками, включая сохранение, получение и проверку хранилища.
type service interface {
	saver
	getter
	Ping(ctx context.Context) error
}
type MetricsServiceServer struct {
	service service
	gen.UnimplementedMetricServiceServer
}

func NewMetricsServiceServer(service service) *MetricsServiceServer {
	return &MetricsServiceServer{service: service}
}

func (s *MetricsServiceServer) mapProtoMetric(metric *gen.Metric) commonmodels.Metric {
	return commonmodels.Metric{
		ID:    metric.Id,
		MType: metric.Type,
		Delta: metric.Delta,
		Value: metric.Value,
	}
}

func (s *MetricsServiceServer) mapCommonMetric(metric commonmodels.Metric) *gen.Metric {
	return &gen.Metric{
		Id:    metric.ID,
		Type:  metric.MType,
		Delta: metric.Delta,
		Value: metric.Value,
	}
}

func (s *MetricsServiceServer) GetAll(ctx context.Context, _ *emptypb.Empty) (*gen.GetAllResponse, error) {
	m, err := s.service.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	mStruct, err := structpb.NewStruct(m)
	if err != nil {
		return nil, err
	}
	return &gen.GetAllResponse{Metrics: mStruct}, nil

}

func (s *MetricsServiceServer) SaveAll(ctx context.Context, in *gen.SaveAllRequest) (*emptypb.Empty, error) {
	metrics := make([]commonmodels.Metric, len(in.Metrics))
	for i, m := range in.Metrics {
		metrics[i] = s.mapProtoMetric(m)
	}
	err := s.service.SaveAll(ctx, metrics)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (s *MetricsServiceServer) Find(ctx context.Context, in *gen.Metric) (*gen.Metric, error) {
	find, err := s.service.Find(ctx, s.mapProtoMetric(in))
	if err != nil {
		return nil, err
	}
	return s.mapCommonMetric(find), nil
}
func (s *MetricsServiceServer) Save(ctx context.Context, in *gen.Metric) (*emptypb.Empty, error) {
	err := s.service.Save(ctx, s.mapProtoMetric(in))
	if err != nil {
		return nil, err
	}
	return nil, nil
}
