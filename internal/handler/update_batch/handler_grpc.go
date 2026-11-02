package updatebatch

import (
	"context"
	"errors"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	pb "github.com/webbash/go-musthave-metrics-tpl.git/internal/proto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler converts protobuf messages without exposing them to the service.
type GRPCHandler struct {
	pb.UnimplementedMetricsServer
	service metricsService
	logger  *zap.SugaredLogger
}

// NewGRPCHandler creates the batch update handler for the gRPC service.
func NewGRPCHandler(service metricsService, logger *zap.SugaredLogger) *GRPCHandler {
	return &GRPCHandler{service: service, logger: logger}
}

func (h *GRPCHandler) UpdateMetrics(ctx context.Context, request *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "missing request")
	}
	metrics := make([]models.Metrics, 0, len(request.GetMetrics()))
	for _, metric := range request.GetMetrics() {
		if metric == nil || metric.GetId() == "" {
			return nil, status.Error(codes.InvalidArgument, "metric name is required")
		}
		value := models.Metrics{ID: metric.GetId()}
		switch metric.GetType() {
		case pb.Metric_COUNTER:
			delta := metric.GetDelta()
			value.MType, value.Delta = models.Counter, &delta
		case pb.Metric_GAUGE:
			gauge := metric.GetValue()
			value.MType, value.Value = models.Gauge, &gauge
		default:
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
		metrics = append(metrics, value)
	}
	if len(metrics) == 0 {
		return &pb.UpdateMetricsResponse{}, nil
	}
	if err := h.service.UpdateMany(ctx, metrics); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, status.FromContextError(err).Err()
		case errors.Is(err, service.ErrInvalidMetric), errors.Is(err, service.ErrInvalidMetricValue), errors.Is(err, service.ErrUnknownMetricType):
			return nil, status.Error(codes.InvalidArgument, "invalid metrics")
		default:
			h.logger.Errorw("failed to save gRPC metrics", "err", err)
			return nil, status.Error(codes.Internal, "failed to save metrics")
		}
	}
	return &pb.UpdateMetricsResponse{}, nil
}
