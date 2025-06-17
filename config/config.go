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
	TLSPrivate   string `env:"CRYPTO_KEY"`
	Address      string `env:"ADDRESS"`
	SaveInterval int    `env:"STORE_INTERVAL"`
	StoragePath  string `env:"FILE_STORAGE_PATH"`
	DBConnection string `env:"DATABASE_DSN"`
	Restore      bool   `env:"RESTORE"`
	HashKey      string `env:"KEY"`
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

type AgentConfig struct {
	PublicKey      string `env:"CRYPTO_KEY"`
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	HashKey        string `env:"KEY"`
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
