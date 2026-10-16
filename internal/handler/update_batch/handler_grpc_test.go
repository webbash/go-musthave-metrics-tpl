package updatebatch

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	pb "github.com/webbash/go-musthave-metrics-tpl.git/internal/proto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type updateFunc func(context.Context, []models.Metrics) error

func (f updateFunc) UpdateMany(ctx context.Context, metrics []models.Metrics) error {
	return f(ctx, metrics)
}

func TestGRPCHandlerValidatesWholeBatchBeforeSaving(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request *pb.UpdateMetricsRequest
	}{
		{"nil request", nil},
		{"nil metric", &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{nil}}},
		{"empty name", &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{{Id: "valid"}, {}}}},
		{"unknown type", &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{{Id: "valid"}, {Id: "invalid", Type: pb.Metric_MType(99)}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			h := NewGRPCHandler(updateFunc(func(context.Context, []models.Metrics) error { calls++; return nil }), zap.NewNop().Sugar())
			_, err := h.UpdateMetrics(t.Context(), tc.request)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.Zero(t, calls)
		})
	}
}

func TestGRPCHandlerPreservesZeroValuesAndMetricTypes(t *testing.T) {
	var got []models.Metrics
	h := NewGRPCHandler(updateFunc(func(_ context.Context, metrics []models.Metrics) error { got = metrics; return nil }), zap.NewNop().Sugar())
	_, err := h.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{
		{Id: "gauge", Delta: 99}, // Default enum is GAUGE; irrelevant delta is ignored.
		{Id: "counter", Type: pb.Metric_COUNTER, Value: 99},
	}})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, models.Gauge, got[0].MType)
	require.NotNil(t, got[0].Value)
	require.Zero(t, *got[0].Value)
	require.Nil(t, got[0].Delta)
	require.Equal(t, models.Counter, got[1].MType)
	require.NotNil(t, got[1].Delta)
	require.Zero(t, *got[1].Delta)
	require.Nil(t, got[1].Value)
}

func TestGRPCHandlerEmptyBatchDoesNotSave(t *testing.T) {
	h := NewGRPCHandler(updateFunc(func(context.Context, []models.Metrics) error {
		t.Error("empty batch must not reach storage")
		return nil
	}), zap.NewNop().Sugar())
	_, err := h.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{})
	require.NoError(t, err)
}

func TestGRPCHandlerMapsServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code codes.Code
	}{
		{errors.New("database password must not leak"), codes.Internal},
		{service.ErrUnknownMetricType, codes.InvalidArgument},
		{context.Canceled, codes.Canceled},
		{context.DeadlineExceeded, codes.DeadlineExceeded},
	} {
		t.Run(tc.code.String(), func(t *testing.T) {
			h := NewGRPCHandler(updateFunc(func(context.Context, []models.Metrics) error { return tc.err }), zap.NewNop().Sugar())
			_, err := h.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{{Id: "metric"}}})
			require.Equal(t, tc.code, status.Code(err))
			require.NotContains(t, err.Error(), "password")
		})
	}
}
