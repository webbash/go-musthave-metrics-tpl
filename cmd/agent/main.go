package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	HashSecret     string `env:"KEY"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	var address string
	var pollInterval int
	var reportInterval int
	var hashSecret string

	flag.StringVar(&address, "a", "localhost:8080", "agent address url")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.StringVar(&hashSecret, "k", "", "hash secret for sending metrics")

	flag.Parse()

	if cfg.Address != "" {
		address = cfg.Address
	}
	if cfg.PollInterval != 0 {
		pollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		reportInterval = cfg.ReportInterval
	}
	if cfg.HashSecret != "" {
		hashSecret = cfg.HashSecret
	}

	var signer *crypto.Sha256Signer
	if hashSecret != "" {
		signer = crypto.NewSha256Signer(hashSecret)
	}

	agentClient := agent.NewAgent(address, time.Duration(pollInterval)*time.Second, time.Duration(reportInterval)*time.Second, &http.Client{}, signer)

	go func() {
		pollTicker := time.NewTicker(agentClient.PollInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-pollTicker.C:
				agentClient.ReadMetrics()
			}
		}
	}()

	reportTicker := time.NewTicker(agentClient.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-reportTicker.C:
			if err := agentClient.SendMetrics(); err != nil {
				slog.Error("failed to send metrics", "error", err)
			}
		}
	}
}
