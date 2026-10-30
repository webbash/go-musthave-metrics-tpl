package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
	updatebatch "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_batch"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	pb "github.com/webbash/go-musthave-metrics-tpl.git/internal/proto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type batchUpdater interface {
	UpdateMany(context.Context, []models.Metrics) error
}

type updateFunc func(context.Context, []models.Metrics) error

func (f updateFunc) UpdateMany(ctx context.Context, metrics []models.Metrics) error {
	return f(ctx, metrics)
}

type testGRPCServer struct {
	*grpcMetricsServer
	certFile, keyFile string
}

func startServer(t *testing.T, address, subnet string, svc batchUpdater) *testGRPCServer {
	t.Helper()
	certFile, keyFile := testTLSFiles(t)
	return startServerWithTLS(t, address, subnet, certFile, keyFile, svc)
}

func startServerWithTLS(t *testing.T, address, subnet, certFile, keyFile string, svc batchUpdater) *testGRPCServer {
	t.Helper()
	s, err := newGRPCMetricsServer(address, subnet, certFile, keyFile, updatebatch.NewGRPCHandler(svc, zap.NewNop().Sugar()))
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() { done <- s.Serve() }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
		require.NoError(t, <-done)
	})
	return &testGRPCServer{grpcMetricsServer: s, certFile: certFile, keyFile: keyFile}
}

func newClient(t *testing.T, s *testGRPCServer) pb.MetricsClient {
	t.Helper()
	creds, err := credentials.NewClientTLSFromFile(s.certFile, "")
	require.NoError(t, err)
	conn, err := grpc.NewClient(s.listener.Addr().String(), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	return pb.NewMetricsClient(conn)
}

func TestSenderIPMatchesConnectionAcrossWorkersAndReconnect(t *testing.T) {
	for _, tc := range []struct{ name, address, subnet string }{
		{"IPv4", "127.0.0.1:0", "127.0.0.0/8"},
		{"IPv6", "[::1]:0", "::1/128"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			addresses := make(chan string, 10)
			svc := updateFunc(func(ctx context.Context, _ []models.Metrics) error {
				p, ok := peer.FromContext(ctx)
				if !ok {
					return status.Error(codes.Internal, "missing peer")
				}
				ip, _, err := net.SplitHostPort(p.Addr.String())
				if err != nil {
					return err
				}
				md, _ := metadata.FromIncomingContext(ctx)
				if values := md.Get("x-real-ip"); len(values) != 1 || values[0] != ip {
					return status.Error(codes.InvalidArgument, "metadata does not match peer")
				}
				addresses <- p.Addr.String()
				return nil
			})
			s := startServer(t, tc.address, tc.subnet, svc)
			address := s.listener.Addr().String()
			sender, err := agent.NewGRPCSender(address, s.certFile)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sender.Close()) })
			value := 0.0
			batch := []models.Metrics{{ID: "zero", MType: models.Gauge, Value: &value}}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			require.NoError(t, sender.Send(ctx, batch))
			require.NoError(t, sender.Send(ctx, batch))
			first, second := <-addresses, <-addresses
			require.Equal(t, first, second, "sequential RPCs reuse one connection")
			var wg sync.WaitGroup
			errs := make(chan error, 4)
			for range 4 {
				wg.Go(func() { errs <- sender.Send(ctx, batch) })
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				require.NoError(t, err)
			}
			for range 4 {
				require.Equal(t, first, <-addresses)
			}
			require.NoError(t, s.Shutdown(ctx))
			startServerWithTLS(t, address, tc.subnet, s.certFile, s.keyFile, svc)
			require.NoError(t, sender.Send(ctx, batch))
			require.NotEqual(t, first, <-addresses, "reconnected RPC uses a new TCP connection")
		})
	}
}

func TestSubnetRejectsRPCBeforeService(t *testing.T) {
	s := startServer(t, "127.0.0.1:0", "127.0.0.0/8", updateFunc(func(context.Context, []models.Metrics) error {
		t.Error("unauthorized RPC reached service")
		return nil
	}))
	client := newClient(t, s)
	_, err := client.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{{Id: "metric"}}})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestShutdownFinishesOrCancelsActiveWrite(t *testing.T) {
	for _, force := range []bool{false, true} {
		name := "graceful"
		if force {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
			s := startServer(t, "127.0.0.1:0", "", updateFunc(func(ctx context.Context, _ []models.Metrics) error {
				close(started)
				defer close(finished)
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}))
			client := newClient(t, s)
			rpcDone := make(chan error, 1)
			go func() {
				_, err := client.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{{Id: "metric"}}})
				rpcDone <- err
			}()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("RPC did not start")
			}
			duration := 2 * time.Second
			if force {
				duration = 30 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(t.Context(), duration)
			defer cancel()
			shutdownDone := make(chan error, 1)
			go func() { shutdownDone <- s.Shutdown(ctx) }()
			if !force {
				select {
				case <-shutdownDone:
					t.Fatal("shutdown returned before write finished")
				case <-time.After(20 * time.Millisecond):
				}
				close(release)
			}
			err := <-shutdownDone
			if force {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			} else {
				require.NoError(t, err)
			}
			select {
			case <-finished:
			default:
				t.Fatal("shutdown returned before handler finished")
			}
			if force {
				require.Error(t, <-rpcDone)
			} else {
				require.NoError(t, <-rpcDone)
			}
		})
	}
}

// Optional integration test; its temporary table cannot modify existing metrics.
func TestPostgresRoundTrip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the PostgreSQL integration test")
	}
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(t.Context(), `CREATE TEMP TABLE metrics (id VARCHAR(200) PRIMARY KEY, type VARCHAR(200) NOT NULL, delta BIGINT, value DOUBLE PRECISION)`)
	require.NoError(t, err)
	repo := repository.NewPostgresRepository(db)
	s := startServer(t, "127.0.0.1:0", "127.0.0.0/8", service.NewMetricsService(repo))
	sender, err := agent.NewGRPCSender(s.listener.Addr().String(), s.certFile)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sender.Close()) })
	delta, value := int64(5), 23.5
	batch := []models.Metrics{{ID: "PollCount", MType: models.Counter, Delta: &delta}, {ID: "temperature", MType: models.Gauge, Value: &value}}
	require.NoError(t, sender.Send(t.Context(), batch))
	value = 0
	require.NoError(t, sender.Send(t.Context(), batch))
	counter, err := repo.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)
	require.EqualValues(t, 10, counter, "current HTTP counter accumulation is preserved")
	gauge, err := repo.GetGauge(t.Context(), "temperature")
	require.NoError(t, err)
	require.Zero(t, gauge)
	client := newClient(t, s)
	ctx := metadata.AppendToOutgoingContext(t.Context(), "x-real-ip", "127.0.0.1")
	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{
		{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 100}, {Id: "bad", Type: pb.Metric_MType(99)},
	}})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	counter, err = repo.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)
	require.EqualValues(t, 10, counter, "invalid batch must not partially update PostgreSQL")
	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{
		{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 100},
		{Id: strings.Repeat("x", 201), Type: pb.Metric_GAUGE, Value: 1},
	}})
	require.Equal(t, codes.Internal, status.Code(err))
	counter, err = repo.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)
	require.EqualValues(t, 10, counter, "database error must roll back the entire batch")
}

// Each test uses its own short-lived certificate, never a checked-in private key.
func testTLSFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)
	cert := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	dir := t.TempDir()
	certFile, keyFile = filepath.Join(dir, "server.crt"), filepath.Join(dir, "server.key")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0600))
	return certFile, keyFile
}
