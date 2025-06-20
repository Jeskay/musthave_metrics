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
	TLSPrivate   string `env:"CRYPTO_KEY" json:"crypto_key"`
	Address      string `env:"ADDRESS" json:"address"`
	SaveInterval int    `env:"STORE_INTERVAL" json:"store_interval"`
	StoragePath  string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DBConnection string `env:"DATABASE_DSN" json:"database_dsn"`
	Restore      bool   `env:"RESTORE" json:"restore"`
	HashKey      string `env:"KEY" json:"key"`
	Config       string `env:"CONFIG"`
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
}

type AgentConfig struct {
	PublicKey      string `env:"CRYPTO_KEY" json:"public_key"`
	Address        string `env:"ADDRESS" json:"address"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
	HashKey        string `env:"KEY" json:"key"`
	Config         string `env:"CONFIG"`
}

func (confFirst *AgentConfig) Merge(confSecond *AgentConfig) {
	if confFirst.Address == "" {
		confFirst.Address = confSecond.Address
	}
	if confFirst.HashKey == "" {
		confFirst.HashKey = confSecond.HashKey
	}
	if confFirst.PublicKey == "" {
		confFirst.PublicKey = confSecond.PublicKey
	}
	if confFirst.PollInterval == -1 {
		confFirst.PollInterval = confSecond.PollInterval
	}
	if confFirst.RateLimit == -1 {
		confFirst.RateLimit = confSecond.RateLimit
	}
	if confFirst.ReportInterval == -1 {
		confFirst.ReportInterval = confSecond.ReportInterval
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
	}
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		Address:        "localhost:8080",
		ReportInterval: 2,
		PollInterval:   10,
		RateLimit:      1,
	}
}
