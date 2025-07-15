// Package sender contains realization of MetricSender interface in forms of HTTP and GRPC clients.
package sender

import (
	"time"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

type MetricSender interface {
	StartSending(interval time.Duration, retrieve func([]string) []dto.Metrics) chan<- struct{}
	CheckAPIAvailability() error
}
