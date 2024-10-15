package internal

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
