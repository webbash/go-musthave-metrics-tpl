package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/logger"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	HashSecret     string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
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
	var rateLimit int

	flag.StringVar(&address, "a", "localhost:8080", "agent address url")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.StringVar(&hashSecret, "k", "", "hash secret for sending metrics")
	flag.IntVar(&rateLimit, "l", 1, "rate limit")

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

	if cfg.RateLimit != 0 {
		rateLimit = cfg.RateLimit
	}

	var signer *crypto.Sha256Signer
	if hashSecret != "" {
		signer = crypto.NewSha256Signer(hashSecret)
	}

	sugar := logger.NewLogger()
	defer sugar.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	agent.NewAgent(
		address,
		time.Duration(pollInterval)*time.Second,
		time.Duration(reportInterval)*time.Second,
		&http.Client{},
		signer,
		rateLimit,
		sugar,
	).Loop(ctx)
}
