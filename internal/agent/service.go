// Package agent contains the business logic and data structures of metric agent.
// The agent sends various memory data over small intervals to the server which collects and stores it.
package agent

import (
	"log/slog"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/agent/monitor"
	"github.com/Jeskay/musthave_metrics/internal/agent/sender"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
)

// AgentService struct provides the functionality of collecting and sending metric data to the server.
type AgentService struct {
	sender        sender.MetricSender
	monitor       *monitor.MetricMonitor
	config        *config.AgentConfig
	JsonAvailable bool
	storage       internal.Repositories
	monitorTick   *time.Ticker
	pollCount     int64
	serverAddr    string
	logger        *slog.Logger
}

// NewAgentService function initializes and returns new instance of AgentService.
func NewAgentService(sender sender.MetricSender, conf *config.AgentConfig, logger slog.Handler) *AgentService {
	service := &AgentService{
		storage:    db.NewMemStorage(),
		serverAddr: "http://" + conf.Address,
		logger:     slog.New(logger),
		config:     conf,
		sender:     sender,
		monitor:    monitor.NewMetricMonitor(logger),
	}
	return service
}

// CheckAPIAvailability of the metric server and returns error if it is unaccessible.
func (svc *AgentService) CheckAPIAvailability() error {
	return svc.sender.CheckAPIAvailability()
}

// StartMonitoring function initiates the process of collecting memory metrics to store in agent's memory storage.
func (svc *AgentService) StartMonitoring(interval time.Duration) chan<- struct{} {
	return svc.monitor.StartMonitoring(interval, svc.storage.SetMany)
}

// StartSending function initiates the process of sending collected data from in-memory storage to the metric server.
// If JSON format is supported by the API, the agent will send metrics in batches thus saving on the amount of requests.
func (svc *AgentService) StartSending(interval time.Duration) chan<- struct{} {
	return svc.sender.StartSending(interval, func(names []string) []dto.Metrics {
		metrics, err := svc.storage.GetMany(names)
		if err != nil {
			svc.logger.Error("failed to retrieve metrics data", slog.Attr{Key: "error", Value: slog.AnyValue(err)})
		}
		return metrics
	})
}
