package internal

import (
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

type Repositories interface {
	Set(key string, value MetricValue)
	Get(key string) (MetricValue, bool)
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

func NewMetric(metric dto.Metrics) *Metric {
	m := &Metric{
		Name: metric.ID,
		Value: MetricValue{
			Type: MetricType(metric.MType),
		},
	}
	if m.Value.Type == CounterMetric {
		m.Value.Value = metric.Delta
	} else {
		m.Value.Value = metric.Value
	}
	return m
}
