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

	if envRestore, ok := os.LookupEnv("RESTORE"); ok {
		if r, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = r
		}
	}

	return cfg
}
