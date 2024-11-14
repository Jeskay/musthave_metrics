package internal

import dto "github.com/Jeskay/musthave_metrics/internal/Dto"

type Repositories interface {
	Set(metric dto.Metrics) error
	SetMany(values []dto.Metrics) error
	Get(key string) (dto.Metrics, bool)
	GetMany(keys []string) ([]dto.Metrics, error)
	Health() bool
	GetAll() []dto.Metrics
}

type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)
