package internal

type Repositories interface {
	Set(key string, value Metric)
	Get(key string) Metric
}

type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Value interface{}
}
