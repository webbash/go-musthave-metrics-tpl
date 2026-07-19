package agent

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type Agent struct {
	PollInterval     time.Duration
	ReportInterval   time.Duration
	RateLimit        int
	sender           *Sender
	runtimeCollector *RuntimeCollector
}

type Batch struct {
	Metrics []models.Metrics
}

func NewAgent(basicURL string, pollInterval, reportInterval time.Duration, httpClient *http.Client, signer *crypto.Sha256Signer, rateLimit int) *Agent {
	addr := basicURL
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return &Agent{
		PollInterval:     pollInterval,
		ReportInterval:   reportInterval,
		RateLimit:        rateLimit,
		sender:           NewSender(httpClient, addr, signer),
		runtimeCollector: NewRuntimeCollector(),
	}
}

func (a *Agent) Loop(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Запускаем горутину для того чтобы собирать метрики из runtime
	go func() {
		pollTicker := time.NewTicker(a.PollInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-pollTicker.C:
				a.runtimeCollector.Collect()
			case <-ctx.Done():
				return
			}
		}
	}()
	// Запускаем горутину для генерации пакетов метрик
	chInput := a.batchesGenerator(ctx)
	wp := NewWorkerPool(a.sender, a.RateLimit, chInput)

	chResult := wp.Start(ctx)

	for result := range chResult {
		fmt.Println(result)
	}
}

func (a *Agent) batchesGenerator(ctx context.Context) <-chan Batch {
	inputCh := make(chan Batch)

	go func() {
		defer close(inputCh)

		pollTicker := time.NewTicker(a.ReportInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-pollTicker.C:
				metrics := a.runtimeCollector.Snapshot()
				inputCh <- Batch{Metrics: metrics}
			case <-ctx.Done():
				return
			}
		}
	}()

	return inputCh
}
