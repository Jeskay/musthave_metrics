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
	if metrics, err := s.storage.GetAll(); err == nil {
		s.file_storage.Save(metrics)
	}
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
			s.storage.Set(m)
		}
	}
}

func (s *MetricService) SetGaugeMetric(key string, value float64) {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %f", key, value))

	s.storage.Set(dto.NewGaugeMetrics(key, value))
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
}

func (s *MetricService) SetCounterMetric(key string, value int64) error {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %d", key, value))

	if v, ok := s.storage.Get(key); ok {
		*v.Delta = *v.Delta + value
		if err := s.storage.Set(v); err != nil {
			s.Logger.Error(err.Error())
			return err
		}
		return nil
	}
	s.storage.Set(dto.NewCounterMetrics(key, value))
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
	return nil
}

func (s *MetricService) SetMetrics(metrics []dto.Metrics) error {
	if err := s.storage.SetMany(metrics); err != nil {
		s.Logger.Error(err.Error())
		return err
	}

	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
	return nil
}

func (s *MetricService) GetMetrics(keys []string) ([]dto.Metrics, error) {
	metrics, err := s.storage.GetMany(keys)
	if err != nil {
		s.Logger.Error(err.Error())
		return nil, err
	}
	return metrics, nil
}

func (s *MetricService) GetCounterMetric(key string) (bool, int64) {
	m, ok := s.storage.Get(key)
	if !ok {
		return false, 0
	}
	if internal.MetricType(m.MType) != internal.CounterMetric || m.Delta == nil {
		return false, 0
	}
	return ok, *m.Delta
}

func (s *MetricService) GetGaugeMetric(key string) (bool, float64) {
	m, ok := s.storage.Get(key)
	if !ok {
		return false, 0
	}
	if internal.MetricType(m.MType) != internal.GaugeMetric || m.Value == nil {
		return false, 0
	}
	return ok, *m.Value
}

func (s *MetricService) GetAllMetrics() ([]dto.Metrics, error) {
	return s.storage.GetAll()
}

func (s *MetricService) DBHealth() bool {
	return s.storage.Health()
}
