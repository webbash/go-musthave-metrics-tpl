package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/logger"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	Transport      string `env:"TRANSPORT"`
	GRPCAddress    string `env:"GRPC_ADDRESS"`
	GRPCTLSCert    string `env:"GRPC_TLS_CERT"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	HashSecret     string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func loadConfig() (Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}

	var address string
	var transport string
	var grpcAddress string
	var grpcTLSCert string
	var pollInterval int
	var reportInterval int
	var hashSecret string
	var cryptoKey string
	var rateLimit int

	flag.StringVar(&address, "a", "localhost:8080", "agent address url")
	flag.StringVar(&transport, "transport", config.HTTP, "metrics transport: http or grpc")
	flag.StringVar(&grpcAddress, "grpc-address", config.DefaultGRPCAddress, "gRPC server endpoint")
	flag.StringVar(&grpcTLSCert, "grpc-tls-cert", "", "path to trusted gRPC TLS certificates")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.StringVar(&hashSecret, "k", "", "hash secret for sending metrics")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to the server public key")
	flag.IntVar(&rateLimit, "l", 1, "rate limit")

	var configPath string
	flag.StringVar(&configPath, "c", "", "path to JSON configuration")
	flag.StringVar(&configPath, "config", "", "path to JSON configuration")
	flag.Parse()

	fileCfg, err := config.ReadAgentFile(configPath)
	if err != nil {
		return Config{}, err
	}
	explicit := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	if fileCfg.Transport != nil && !explicit["transport"] {
		transport = *fileCfg.Transport
	}
	if fileCfg.GRPCAddress != nil && !explicit["grpc-address"] {
		grpcAddress = *fileCfg.GRPCAddress
	}
	if fileCfg.GRPCTLSCert != nil && !explicit["grpc-tls-cert"] {
		grpcTLSCert = *fileCfg.GRPCTLSCert
	}
	if fileCfg.Address != nil && !explicit["a"] {
		address = *fileCfg.Address
	}
	if fileCfg.PollInterval != nil && !explicit["p"] {
		pollInterval = int(*fileCfg.PollInterval)
	}
	if fileCfg.ReportInterval != nil && !explicit["r"] {
		reportInterval = int(*fileCfg.ReportInterval)
	}
	if fileCfg.HashSecret != nil && !explicit["k"] {
		hashSecret = *fileCfg.HashSecret
	}
	if fileCfg.CryptoKey != nil && !explicit["crypto-key"] {
		cryptoKey = *fileCfg.CryptoKey
	}
	if fileCfg.RateLimit != nil && !explicit["l"] {
		rateLimit = *fileCfg.RateLimit
	}

	if cfg.Address != "" {
		address = cfg.Address
	}
	if cfg.Transport != "" {
		transport = cfg.Transport
	}
	if cfg.GRPCAddress != "" {
		grpcAddress = cfg.GRPCAddress
	}
	if cfg.GRPCTLSCert != "" {
		grpcTLSCert = cfg.GRPCTLSCert
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
	if cfg.CryptoKey != "" {
		cryptoKey = cfg.CryptoKey
	}

	if cfg.RateLimit != 0 {
		rateLimit = cfg.RateLimit
	}

	return Config{Address: address, Transport: transport, GRPCAddress: grpcAddress, GRPCTLSCert: grpcTLSCert,
		PollInterval: pollInterval, ReportInterval: reportInterval,
		HashSecret: hashSecret, CryptoKey: cryptoKey, RateLimit: rateLimit}, nil
}

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := config.ValidateTransport(cfg.Transport); err != nil {
		log.Fatal(err)
	}

	sugar := logger.NewLogger()
	defer sugar.Sync()

	collectCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	sendCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var a *agent.Agent
	if cfg.Transport == config.GRPC {
		if cfg.GRPCAddress == "" {
			log.Fatal("gRPC address must not be empty")
		}
		sender, err := agent.NewGRPCSender(cfg.GRPCAddress, cfg.GRPCTLSCert)
		if err != nil {
			log.Fatal(err)
		}
		defer sender.Close()
		a = agent.NewAgentWithSender(sender, time.Duration(cfg.PollInterval)*time.Second,
			time.Duration(cfg.ReportInterval)*time.Second, cfg.RateLimit, sugar)
	} else {
		var signer *crypto.SHA256Signer
		if cfg.HashSecret != "" {
			signer = crypto.NewSHA256Signer(cfg.HashSecret)
		}
		var encryptor *crypto.Encryptor
		if cfg.CryptoKey != "" {
			publicKey, err := crypto.LoadPublicKey(cfg.CryptoKey)
			if err != nil {
				log.Fatalf("load public crypto key: %v", err)
			}
			encryptor = crypto.NewEncryptor(publicKey)
		}
		a = agent.NewAgent(cfg.Address, time.Duration(cfg.PollInterval)*time.Second,
			time.Duration(cfg.ReportInterval)*time.Second, &http.Client{}, signer, cfg.RateLimit, sugar, encryptor)
	}
	stopBatches := make(chan struct{})
	collectors, workers := a.Start(collectCtx, sendCtx, stopBatches)

	<-collectCtx.Done()
	collectors.Wait()
	close(stopBatches)
	workers.Wait()
	if err := a.SendSnapshot(sendCtx); err != nil {
		sugar.Errorw("failed to send final metrics", "err", err)
	}
}
