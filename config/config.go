package config

import "time"

type ServerConfig struct {
	Address      string `env:"ADDRESS"`
	SaveInterval int    `env:"STORE_INTERVAL"`
	StoragePath  string `env:"FILE_STORAGE_PATH"`
	DBConnection string `env:"DATABASE_DSN"`
	Restore      bool   `env:"RESTORE"`
}

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
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
	}
}
