package grpc

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/protos/gen"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"testing"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Save(ctx context.Context, metric models.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *mockService) SaveAll(ctx context.Context, metrics []models.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockService) Find(ctx context.Context, metric models.Metric) (models.Metric, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(models.Metric), args.Error(1)
}

func (m *mockService) GetAll(ctx context.Context) (map[string]any, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *mockService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewMetricsServiceServer(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	assert.NotNil(t, server)
	assert.Equal(t, svc, server.service)
}

func TestMapProtoMetric(t *testing.T) {
	server := &MetricsServiceServer{}

	delta := utils.MakePointer[int64](100)
	value := utils.MakePointer[float64](42.5)

	protoMetric := &gen.Metric{
		Id:    "test_id",
		Type:  "gauge",
		Delta: delta,
		Value: value,
	}
	expected := models.Metric{
		ID:    "test_id",
		MType: "gauge",
		Delta: delta,
		Value: value,
	}
	result := server.mapProtoMetric(protoMetric)
	assert.Equal(t, expected, result)
}

func TestMapCommonMetric(t *testing.T) {
	server := &MetricsServiceServer{}

	delta := utils.MakePointer[int64](200)
	value := utils.MakePointer[float64](42.5)

	commonMetric := models.Metric{
		ID:    "test_id",
		MType: "counter",
		Delta: delta,
		Value: value,
	}
	expected := &gen.Metric{
		Id:    "test_id",
		Type:  "counter",
		Delta: delta,
		Value: value,
	}
	result := server.mapCommonMetric(commonMetric)
	assert.Equal(t, expected, result)
}

func TestGetAllSuccess(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)

	ctx := context.Background()
	data := map[string]any{
		"metric1": map[string]any{"id": "metric1", "type": "gauge", "value": 42.5},
	}
	svc.On("GetAll", ctx).Return(data, nil)

	resp, err := server.GetAll(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	assert.NotNil(t, resp.Metrics)

	mStruct, err := structpb.NewStruct(data)
	require.NoError(t, err)
	assert.Equal(t, mStruct, resp.Metrics)
	svc.AssertExpectations(t)
}

func TestGetAllError(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)

	ctx := context.Background()
	svc.On("GetAll", ctx).Return(map[string]any{}, errors.New("get all error"))

	resp, err := server.GetAll(ctx, &emptypb.Empty{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	svc.AssertExpectations(t)
}

func TestGetAllInvalidStruct(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)

	ctx := context.Background()
	data := map[string]any{
		"metric1": complex(1, 2),
	}
	svc.On("GetAll", ctx).Return(data, nil)

	resp, err := server.GetAll(ctx, &emptypb.Empty{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	svc.AssertExpectations(t)
}

func TestSaveAllSuccess(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)
	delta := utils.MakePointer[int64](100)

	ctx := context.Background()
	in := &gen.SaveAllRequest{
		Metrics: []*gen.Metric{
			{Id: "metric1", Type: "gauge", Value: value},
			{Id: "metric2", Type: "counter", Delta: delta},
		},
	}
	metrics := []models.Metric{
		{ID: "metric1", MType: "gauge", Value: value},
		{ID: "metric2", MType: "counter", Delta: delta},
	}
	svc.On("SaveAll", ctx, metrics).Return(nil)

	_, err := server.SaveAll(ctx, in)
	require.NoError(t, err)
	svc.AssertExpectations(t)
}

func TestSaveAllError(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)

	ctx := context.Background()
	in := &gen.SaveAllRequest{
		Metrics: []*gen.Metric{
			{Id: "metric1", Type: "gauge", Value: value},
		},
	}
	metrics := []models.Metric{
		{ID: "metric1", MType: "gauge", Value: value},
	}
	svc.On("SaveAll", ctx, metrics).Return(errors.New("save all error"))

	resp, err := server.SaveAll(ctx, in)
	assert.Error(t, err)
	assert.Nil(t, resp)
	svc.AssertExpectations(t)
}

func TestFindSuccess(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)

	ctx := context.Background()
	in := &gen.Metric{Id: "metric1", Type: "gauge", Value: value}
	commonMetric := models.Metric{ID: "metric1", MType: "gauge", Value: value}
	svc.On("Find", ctx, commonMetric).Return(commonMetric, nil)

	resp, err := server.Find(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, in, resp)
	svc.AssertExpectations(t)
}

func TestFindError(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)

	ctx := context.Background()
	in := &gen.Metric{Id: "metric1", Type: "gauge", Value: value}
	commonMetric := models.Metric{ID: "metric1", MType: "gauge", Value: value}
	svc.On("Find", ctx, commonMetric).Return(models.Metric{}, errors.New("find error"))

	resp, err := server.Find(ctx, in)
	assert.Error(t, err)
	assert.Nil(t, resp)
	svc.AssertExpectations(t)
}

func TestSaveSuccess(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)

	ctx := context.Background()
	in := &gen.Metric{Id: "metric1", Type: "gauge", Value: value}
	commonMetric := models.Metric{ID: "metric1", MType: "gauge", Value: value}
	svc.On("Save", ctx, commonMetric).Return(nil)

	_, err := server.Save(ctx, in)
	require.NoError(t, err)
	svc.AssertExpectations(t)
}

func TestSaveError(t *testing.T) {
	svc := &mockService{}
	server := NewMetricsServiceServer(svc)
	value := utils.MakePointer[float64](42.5)

	ctx := context.Background()
	in := &gen.Metric{Id: "metric1", Type: "gauge", Value: value}
	commonMetric := models.Metric{ID: "metric1", MType: "gauge", Value: value}
	svc.On("Save", ctx, commonMetric).Return(errors.New("save error"))

	resp, err := server.Save(ctx, in)
	assert.Error(t, err)
	assert.Nil(t, resp)
	svc.AssertExpectations(t)
}
