package config

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address: "localhost:8080",
	}
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		Address:        "localhost:8080",
		ReportInterval: 2,
		PollInterval:   10,
	}
}
