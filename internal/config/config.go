package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	Restore         bool
	StoreInterval   int
	FileStoragePath string
}

func NewConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "metric collection to file interval")
	flag.StringVar(&cfg.FileStoragePath, "f", "./tmp/temporary.json", "file memRepository path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")

	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if seconds, err := strconv.Atoi(envStoreInterval); err == nil {
			cfg.StoreInterval = seconds
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if r, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = r
		}
	}

	return cfg
}
