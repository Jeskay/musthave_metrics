package metric

import "github.com/Jeskay/musthave_metrics/internal"

type MetricService struct {
	storage *internal.MemStorage
}

func NewMetricService() *MetricService {
	service := &MetricService{
		storage: internal.NewMemStorage(),
	}
	return service
}

func (s *MetricService) SetGaugeMetric(key string, value float64) {
	s.storage.Set(key, internal.Metric{Type: internal.GaugeMetric, Value: value})
}

func (s *MetricService) SetCounterMetric(key string, value int64) {
	s.storage.Set(key, internal.Metric{Type: internal.CounterMetric, Value: value})
}
