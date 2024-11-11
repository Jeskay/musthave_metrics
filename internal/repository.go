package internal

import (
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

type Repositories interface {
	Set(key string, value MetricValue) error
	SetMany(values []Metric) error
	Get(key string) (MetricValue, bool)
	GetMany(keys []string) ([]*Metric, error)
	Health() bool
	GetAll() []*Metric
}

type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)

type MetricValue struct {
	Type  MetricType
	Value interface{}
}

type Metric struct {
	Name  string
	Value MetricValue
}

func (m Metric) ToDto() dto.Metrics {
	d := dto.Metrics{
		ID:    m.Name,
		MType: string(m.Value.Type),
	}
	if v, ok := m.Value.Value.(int64); ok {
		d.Delta = &v
	} else if v, ok := m.Value.Value.(float64); ok {
		d.Value = &v
	}
	return d
}

func NewMetric(metric dto.Metrics) *Metric {
	m := &Metric{
		Name: metric.ID,
		Value: MetricValue{
			Type: MetricType(metric.MType),
		},
	}
	if m.Value.Type == CounterMetric {
		m.Value.Value = *metric.Delta
	} else {
		m.Value.Value = *metric.Value
	}
	return m
}
