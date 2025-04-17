package config

import "time"

type ServerConfig struct {
	Address      string `env:"ADDRESS"`
	SaveInterval int    `env:"STORE_INTERVAL"`
	StoragePath  string `env:"FILE_STORAGE_PATH"`
	DBConnection string `env:"DATABASE_DSN"`
	Restore      bool   `env:"RESTORE"`
	HashKey      string `env:"KEY"`
}

type AgentConfig struct {
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
