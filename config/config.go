package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"time"
)

type ServerConfig struct {
	TLSPrivate    string `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	Address       string `env:"ADDRESS" json:"address"`
	SaveInterval  int    `env:"STORE_INTERVAL" json:"store_interval"`
	StoragePath   string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DBConnection  string `env:"DATABASE_DSN" json:"database_dsn"`
	Restore       bool   `env:"RESTORE" json:"restore"`
	GRPC          bool   `env:"GRPC" json:"grpc"`
	HashKey       string `env:"KEY" json:"key"`
	Config        string `env:"CONFIG"`
}

func (cfg *ServerConfig) LoadPrivateKey() (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(cfg.TLSPrivate)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("failed to load private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func (cfg *ServerConfig) Merge(cfgMerge *ServerConfig) {
	if cfg.TLSPrivate == "" {
		cfg.TLSPrivate = cfgMerge.TLSPrivate
	}
	if cfg.Address == "" {
		cfg.Address = cfgMerge.Address
	}
	if cfg.SaveInterval == 300 {
		cfg.SaveInterval = cfgMerge.SaveInterval
	}
	if cfg.StoragePath == "" {
		cfg.StoragePath = cfgMerge.StoragePath
	}
	if cfg.DBConnection == "" {
		cfg.DBConnection = cfgMerge.DBConnection
	}
	cfg.Restore = cfgMerge.Restore
	if cfg.HashKey == "" {
		cfg.HashKey = cfgMerge.HashKey
	}
	if cfg.Config == "" {
		cfg.Config = cfgMerge.Config
	}
	if !cfg.GRPC {
		cfg.GRPC = cfgMerge.GRPC
	}
}

type AgentConfig struct {
	PublicKey      string `env:"CRYPTO_KEY" json:"public_key"`
	Address        string `env:"ADDRESS" json:"address"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
	HashKey        string `env:"KEY" json:"key"`
	Config         string `env:"CONFIG"`
	GRPC           bool   `env:"GRPC" json:"grpc"`
}

func (cfg *AgentConfig) Merge(cfgMerge *AgentConfig) {
	if cfg.Address == "" {
		cfg.Address = cfgMerge.Address
	}
	if cfg.HashKey == "" {
		cfg.HashKey = cfgMerge.HashKey
	}
	if cfg.PublicKey == "" {
		cfg.PublicKey = cfgMerge.PublicKey
	}
	if cfg.PollInterval == -1 {
		cfg.PollInterval = cfgMerge.PollInterval
	}
	if cfg.RateLimit == -1 {
		cfg.RateLimit = cfgMerge.RateLimit
	}
	if cfg.ReportInterval == -1 {
		cfg.ReportInterval = cfgMerge.ReportInterval
	}
	if !cfg.GRPC {
		cfg.GRPC = cfgMerge.GRPC
	}
}

func (cfg *AgentConfig) GetReportInterval() time.Duration {
	return time.Second * time.Duration(cfg.ReportInterval)
}

func (cfg *AgentConfig) GetPollInterval() time.Duration {
	return time.Second * time.Duration(cfg.PollInterval)
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:      "localhost:8080",
		SaveInterval: 300,
		StoragePath:  "/metrics.dat",
		Restore:      true,
		GRPC:         false,
	}
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		Address:        "localhost:8080",
		ReportInterval: 2,
		PollInterval:   10,
		RateLimit:      1,
		GRPC:           false,
	}
}
