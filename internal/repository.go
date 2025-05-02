// Package internal contains structures and variables related to application business logic.
package internal

import dto "github.com/Jeskay/musthave_metrics/internal/Dto"

// Repositories represents metrics storage functionality.
type Repositories interface {
	Set(metric dto.Metrics) error
	SetMany(values []dto.Metrics) error
	Get(key string) (dto.Metrics, bool)
	GetMany(keys []string) ([]dto.Metrics, error)
	Health() bool
	GetAll() ([]dto.Metrics, error)
}

type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)

const HashHeader = "HashSHA256"
