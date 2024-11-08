package metric

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
)

type MetricService struct {
	storage      internal.Repositories
	file_storage *db.FileStorage
	Logger       *slog.Logger
	conf         config.ServerConfig
	ticker       *time.Ticker
	close        chan struct{}
}

func NewMetricService(conf config.ServerConfig, logger slog.Handler, file_storage *db.FileStorage, memory_storage internal.Repositories) *MetricService {
	service := &MetricService{
		storage:      memory_storage,
		file_storage: file_storage,
		Logger:       slog.New(logger),
		conf:         conf,
		close:        make(chan struct{}),
	}
	if !service.databaseAccessible() {
		service.LoadSavings()
		go service.StartSaving()
	}
	return service
}

func (s *MetricService) shouldSaveInstantly() bool {
	return s.conf.SaveInterval == 0
}

func (s *MetricService) databaseAccessible() bool {
	return s.conf.DBConnection != ""
}

func (s *MetricService) saveMetrics() {
	m := []internal.Metric{}
	for _, metric := range s.storage.GetAll() {
		m = append(m, *metric)
	}
	s.file_storage.Save(m)
}

func (s *MetricService) Close() {
	if !s.databaseAccessible() {
		s.close <- struct{}{}
	}
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

func (s *MetricService) LoadSavings() {
	if metrics, err := s.file_storage.Load(); err == nil {
		for _, m := range metrics {
			s.storage.Set(m.Name, m.Value)
		}
	}
}

func (s *MetricService) SetGaugeMetric(key string, value float64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %f", key, value))

	s.storage.Set(key, internal.MetricValue{Type: internal.GaugeMetric, Value: value})
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
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
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
}

func (s *MetricService) SetMetrics(metrics []dto.Metrics) {
	cMetrics := make([]internal.Metric, 0)
	for _, v := range metrics {
		cMetrics = append(cMetrics, *internal.NewMetric(v))
	}
	s.storage.SetMany(cMetrics)
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
}

func (s *MetricService) GetMetrics(keys []string) []dto.Metrics {
	metrics := s.storage.GetMany(keys)
	ms := make([]dto.Metrics, len(metrics))
	for i, m := range metrics {
		ms[i] = m.ToDto()
	}
	return ms
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

func (s *MetricService) DBHealth() bool {
	return s.storage.Health()
}
