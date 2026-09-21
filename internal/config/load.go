package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ServerFile JSON fields are pointers so omitted values differ from false, 0 and "".
type ServerFile struct {
	Address       *string  `json:"address"`
	Restore       *bool    `json:"restore"`
	StoreInterval *Seconds `json:"store_interval"`
	StoreFile     *string  `json:"store_file"`
	DatabaseDSN   *string  `json:"database_dsn"`
	HashSecret    *string  `json:"key"`
	CryptoKey     *string  `json:"crypto_key"`
	AuditFile     *string  `json:"audit_file"`
	AuditURL      *string  `json:"audit_url"`
}

type AgentFile struct {
	Address        *string  `json:"address"`
	PollInterval   *Seconds `json:"poll_interval"`
	ReportInterval *Seconds `json:"report_interval"`
	HashSecret     *string  `json:"key"`
	CryptoKey      *string  `json:"crypto_key"`
	RateLimit      *int     `json:"rate_limit"`
}

// Seconds decodes a JSON duration string into the existing integer-seconds format.
type Seconds int

func (s *Seconds) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	if duration < 0 || duration%time.Second != 0 {
		return fmt.Errorf("interval must be a non-negative whole number of seconds")
	}
	*s = Seconds(duration / time.Second)
	return nil
}

func ReadServerFile(path string) (ServerFile, error) {
	if value, ok := os.LookupEnv("CONFIG"); ok {
		path = value
	}
	if path == "" {
		return ServerFile{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ServerFile{}, fmt.Errorf("read server configuration: %w", err)
	}
	var cfg *ServerFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ServerFile{}, fmt.Errorf("decode server configuration: %w", err)
	}
	if cfg == nil {
		return ServerFile{}, fmt.Errorf("server configuration must be a JSON object")
	}
	return *cfg, nil
}

func ReadAgentFile(path string) (AgentFile, error) {
	if value, ok := os.LookupEnv("CONFIG"); ok {
		path = value
	}
	if path == "" {
		return AgentFile{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentFile{}, fmt.Errorf("read agent configuration: %w", err)
	}
	var cfg *AgentFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AgentFile{}, fmt.Errorf("decode agent configuration: %w", err)
	}
	if cfg == nil {
		return AgentFile{}, fmt.Errorf("agent configuration must be a JSON object")
	}
	if cfg.PollInterval != nil && *cfg.PollInterval == 0 {
		return AgentFile{}, fmt.Errorf("poll_interval must be positive")
	}
	if cfg.ReportInterval != nil && *cfg.ReportInterval == 0 {
		return AgentFile{}, fmt.Errorf("report_interval must be positive")
	}
	return *cfg, nil
}
