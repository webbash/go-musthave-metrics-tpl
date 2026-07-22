package agent

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"go.uber.org/zap"
)

type Agent struct {
	PollInterval      time.Duration
	ReportInterval    time.Duration
	RateLimit         int
	sender            *Sender
	runtimeCollector  *RuntimeCollector
	gopsutilCollector *GopsutilCollector
	logger            *zap.SugaredLogger
}

type Batch struct {
	Metrics []models.Metrics
}

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
