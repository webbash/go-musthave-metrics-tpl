// Package agent collects runtime and system metrics and sends them to the
// metrics server in batches.
package agent

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

// Agent coordinates metric collection and delivery.
type Agent struct {
	// PollInterval is the interval at which collectors refresh their snapshots.
	PollInterval time.Duration
	// ReportInterval is the interval at which collected metrics are grouped into batches.
	ReportInterval time.Duration
	// RateLimit is the maximum number of concurrent workers sending batches.
	RateLimit         int
	sender            *Sender
	runtimeCollector  *RuntimeCollector
	gopsutilCollector *GopsutilCollector
	logger            *zap.SugaredLogger
}

// Batch contains the metrics sent by one worker-pool job.
type Batch struct {
	// Metrics is the collection of metrics to send to the server.
	Metrics []models.Metrics
}

// NewAgent creates an agent configured to send metrics to basicURL.
// If basicURL has no scheme, http:// is used.
func NewAgent(basicURL string, pollInterval, reportInterval time.Duration, httpClient *http.Client, signer *crypto.SHA256Signer, rateLimit int, logger *zap.SugaredLogger) *Agent {
	addr := basicURL
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return &Agent{
		PollInterval:      pollInterval,
		ReportInterval:    reportInterval,
		RateLimit:         rateLimit,
		sender:            NewSender(httpClient, addr, signer),
		logger:            logger,
		runtimeCollector:  NewRuntimeCollector(),
		gopsutilCollector: NewGopsutilCollector(),
	}
}

// Loop starts metric collection and sending until ctx is cancelled.
func (a *Agent) Loop(ctx context.Context) {
	wg := sync.WaitGroup{}

	// Запускаем горутину для того чтобы собирать метрики из runtime
	wg.Add(1)
	go func() {
		pollTicker := time.NewTicker(a.PollInterval)
		defer pollTicker.Stop()
		defer func() {
			a.logger.Infow("runtime collector has finished")
			wg.Done()
		}()

		for {
			select {
			case <-pollTicker.C:
				a.runtimeCollector.Collect()
			case <-ctx.Done():
				return
			}
		}
	}()
	// Запускаем горутину для того чтобы собирать метрики из gopsutil
	wg.Add(1)
	go func() {
		pollTicker := time.NewTicker(a.PollInterval)
		defer pollTicker.Stop()
		defer func() {
			a.logger.Infow("gopsutil collector has finished")
			wg.Done()
		}()

		for {
			select {
			case <-pollTicker.C:
				err := a.gopsutilCollector.Collect()
				if err != nil {
					a.logger.Errorw("failed to get gopsutil metrics", "err", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Запускаем горутину для генерации пакетов метрик
	chInput := a.batchesGenerator(ctx)
	wp := NewWorkerPool(a.sender, a.RateLimit, chInput, a.logger)

	wp.Start(ctx)

	wp.Wait()
	wg.Wait()
}

func (a *Agent) batchesGenerator(ctx context.Context) <-chan Batch {
	inputCh := make(chan Batch)

	go func() {
		defer func() {
			close(inputCh)
			a.logger.Infow("channel inputCh closing...")
		}()

		pollTicker := time.NewTicker(a.ReportInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-pollTicker.C:
				runtimeMetrics := a.runtimeCollector.Snapshot()
				systemMetrics := a.gopsutilCollector.Snapshot()

				batch := Batch{Metrics: append(runtimeMetrics, systemMetrics...)}

				select {
				case inputCh <- batch:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return inputCh
}
