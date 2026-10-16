package agent

import (
	"context"
	"fmt"
	"time"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	pb "github.com/webbash/go-musthave-metrics-tpl.git/internal/proto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCSender shares one client connection between concurrent workers.
type GRPCSender struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
}

// NewGRPCSender creates a plaintext client with connection-specific IP metadata.
func NewGRPCSender(address string) (*GRPCSender, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(newLocalIPTransport()),
		grpc.WithPerRPCCredentials(localIPMetadata{}),
		grpc.WithDisableRetry(),
	)
	if err != nil {
		return nil, fmt.Errorf("create gRPC client: %w", err)
	}
	return &GRPCSender{conn: conn, client: pb.NewMetricsClient(conn)}, nil
}

// Close releases the connection after workers and the final send have finished.
func (s *GRPCSender) Close() error { return s.conn.Close() }

// Send preserves the batch across retries and only retries Unavailable.
func (s *GRPCSender) Send(ctx context.Context, metrics []models.Metrics) error {
	request := &pb.UpdateMetricsRequest{Metrics: make([]*pb.Metric, 0, len(metrics))}
	for _, metric := range metrics {
		value := &pb.Metric{Id: metric.ID}
		switch metric.MType {
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter %q has no delta", metric.ID)
			}
			value.Type = pb.Metric_COUNTER
			value.Delta = *metric.Delta
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge %q has no value", metric.ID)
			}
			value.Type = pb.Metric_GAUGE
			value.Value = *metric.Value
		default:
			return fmt.Errorf("unknown metric type %q", metric.MType)
		}
		request.Metrics = append(request.Metrics, value)
	}
	err := retry.Do(ctx, func() error {
		attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err := s.client.UpdateMetrics(attemptCtx, request)
		return err
	}, func(err error) bool {
		return status.Code(err) == codes.Unavailable
	}, []time.Duration{time.Second, 3 * time.Second, 5 * time.Second})
	if err != nil {
		return fmt.Errorf("send gRPC metrics: %w", err)
	}
	return nil
}
