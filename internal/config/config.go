// Package config loads the application's configuration.
package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	Restore         bool
	StoreInterval   int
	FileStoragePath string
	DatabaseDSN     string
	HashSecret      string
	CryptoKey       string
	AuditFile       string
	AuditURL        string
}

func NewConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "metric collection to file interval")
	flag.StringVar(&cfg.FileStoragePath, "f", "./tmp/temporary.json", "file memRepository path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "DB connection string")
	flag.StringVar(&cfg.HashSecret, "k", "", "Hash secret for receiving metrics")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to the server private key")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Path to file for audit")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Url for audit")

	var configPath string
	flag.StringVar(&configPath, "c", "", "path to JSON configuration")
	flag.StringVar(&configPath, "config", "", "path to JSON configuration")
	flag.Parse()

	fileCfg, err := ReadServerFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	explicit := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	if fileCfg.Address != nil && !explicit["a"] {
		cfg.Address = *fileCfg.Address
	}
	if fileCfg.Restore != nil && !explicit["r"] {
		cfg.Restore = *fileCfg.Restore
	}
	if fileCfg.StoreInterval != nil && !explicit["i"] {
		cfg.StoreInterval = int(*fileCfg.StoreInterval)
	}
	if fileCfg.StoreFile != nil && !explicit["f"] {
		cfg.FileStoragePath = *fileCfg.StoreFile
	}
	if fileCfg.DatabaseDSN != nil && !explicit["d"] {
		cfg.DatabaseDSN = *fileCfg.DatabaseDSN
	}
	if fileCfg.HashSecret != nil && !explicit["k"] {
		cfg.HashSecret = *fileCfg.HashSecret
	}
	if fileCfg.CryptoKey != nil && !explicit["crypto-key"] {
		cfg.CryptoKey = *fileCfg.CryptoKey
	}
	if fileCfg.AuditFile != nil && !explicit["audit-file"] {
		cfg.AuditFile = *fileCfg.AuditFile
	}
	if fileCfg.AuditURL != nil && !explicit["audit-url"] {
		cfg.AuditURL = *fileCfg.AuditURL
	}

	if envAddress, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Address = envAddress
	}

	if envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if seconds, err := strconv.Atoi(envStoreInterval); err == nil {
			cfg.StoreInterval = seconds
		}
	}

	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = envDsn
	}

	if envRestore, ok := os.LookupEnv("RESTORE"); ok {
		if r, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = r
		}
	}

	if envSecret, ok := os.LookupEnv("KEY"); ok {
		cfg.HashSecret = envSecret
	}

	if envCryptoKey, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cfg.CryptoKey = envCryptoKey
	}

	if envAuditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = envAuditURL
	}

	return cfg
}
