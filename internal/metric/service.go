package metric

import (
	"fmt"
	"log/slog"

	"github.com/Jeskay/musthave_metrics/internal"
)

type MetricService struct {
	storage *internal.MemStorage
	Logger  *slog.Logger
}

func NewMetricService(logger slog.Handler) *MetricService {
	service := &MetricService{
		storage: internal.NewMemStorage(),
		Logger:  slog.New(logger),
	}
	return service
}

func (s *MetricService) SetGaugeMetric(key string, value float64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %f", key, value))

	s.storage.Set(key, internal.MetricValue{Type: internal.GaugeMetric, Value: value})
}

func (s *MetricService) SetCounterMetric(key string, value int64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %d", key, value))

	v, ok := s.storage.Get(key)
	if ok {
		if old, ok := v.Value.(int64); ok {
			v.Value = old + value
			s.storage.Set(key, v)
			return
		}
	}
	s.storage.Set(key, internal.MetricValue{Type: internal.CounterMetric, Value: value})
}

func (s *MetricService) GetCounterMetric(key string) (bool, int64) {
	m, ok := s.storage.Get(key)
	if !ok {
		return false, 0
	}
	if m.Type != internal.CounterMetric {
		return false, 0
	}
	value, ok := m.Value.(int64)
	return ok, value
}

func (s *MetricService) GetGaugeMetric(key string) (bool, float64) {
	m, ok := s.storage.Get(key)
	if !ok {
		return false, 0
	}
	if m.Type != internal.GaugeMetric {
		return false, 0
	}
	value, ok := m.Value.(float64)
	return ok, value
}

func (s *MetricService) GetAllMetrics() []*internal.Metric {
	return s.storage.GetAll()
}
