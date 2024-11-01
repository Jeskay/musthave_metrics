package metric

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
)

type MetricService struct {
	memory_storage *internal.MemStorage
	file_storage   *internal.FileStorage
	Logger         *slog.Logger
	conf           config.ServerConfig
	ticker         *time.Ticker
	close          chan struct{}
}

func NewMetricService(conf config.ServerConfig, logger slog.Handler, file_storage *internal.FileStorage, memory_storage *internal.MemStorage) *MetricService {
	service := &MetricService{
		memory_storage: memory_storage,
		file_storage:   file_storage,
		Logger:         slog.New(logger),
		conf:           conf,
		close:          make(chan struct{}),
	}
	if metrics, err := service.file_storage.Load(); err == nil {
		for _, m := range metrics {
			service.memory_storage.Set(m.Name, m.Value)
		}
	}
	go service.StartSaving()
	return service
}

func (s *MetricService) shouldSaveInstantly() bool {
	return s.conf.SaveInterval == 0
}

func (s *MetricService) saveMetrics() {
	m := []internal.Metric{}
	for _, metric := range s.memory_storage.GetAll() {
		m = append(m, *metric)
	}
	s.file_storage.Save(m)
}

func (s *MetricService) Close() {
	s.close <- struct{}{}
}

func (s *MetricService) StartSaving() {
	s.ticker = time.NewTicker(time.Duration(s.conf.SaveInterval) * time.Second)
	go func() {
		for {
			select {
			case <-s.close:
				s.ticker.Stop()
				close(s.close)
				return
			case <-s.ticker.C:
				s.saveMetrics()
			}
		}
	}()
}

func (s *MetricService) SetGaugeMetric(key string, value float64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %f", key, value))

	s.memory_storage.Set(key, internal.MetricValue{Type: internal.GaugeMetric, Value: value})
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
}

func (s *MetricService) SetCounterMetric(key string, value int64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %d", key, value))

	v, ok := s.memory_storage.Get(key)
	if ok {
		if old, ok := v.Value.(int64); ok {
			v.Value = old + value
			s.memory_storage.Set(key, v)
			return
		}
	}
	s.memory_storage.Set(key, internal.MetricValue{Type: internal.CounterMetric, Value: value})
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
}

func (s *MetricService) GetCounterMetric(key string) (bool, int64) {
	m, ok := s.memory_storage.Get(key)
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
	m, ok := s.memory_storage.Get(key)
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
	return s.memory_storage.GetAll()
}
