// Package metric contains functionality of the server that stores and
// updates metric data, sent by the agent.
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

// MetricService represents the service for storing and updating metric data.
type MetricService struct {
	storage     internal.Repositories
	fileStorage *db.FileStorage
	Logger      *slog.Logger // Instance of logger to write error and debug information to.
	conf        config.ServerConfig
	ticker      *time.Ticker
	close       chan struct{}
}

// NewMetricService function initialize and returns new MetricService instance.
// The function also loads previously saved metric data from local storage if database is unaccessible.
func NewMetricService(conf config.ServerConfig, logger slog.Handler, fileStorage *db.FileStorage, memoryStorage internal.Repositories) *MetricService {
	service := &MetricService{
		storage:     memoryStorage,
		fileStorage: fileStorage,
		Logger:      slog.New(logger),
		conf:        conf,
		close:       make(chan struct{}),
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
		s.fileStorage.Save(metrics)
	}
}

// Close function initiates the stop of metric saving goroutine.
func (s *MetricService) Close() {
	if !s.databaseAccessible() {
		s.close <- struct{}{}
	}
}

// StartSaving function starts metric saving goroutine.
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

// LoadSavings function loads metric data from file storage to the database.
func (s *MetricService) LoadSavings() {
	if metrics, err := s.fileStorage.Load(); err == nil {
		for _, m := range metrics {
			s.storage.Set(m)
		}
	}
}

// SetGaugeMetric function sets gauge metric to the specified value.
func (s *MetricService) SetGaugeMetric(key string, value float64) error {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %f", key, value))

	err := s.storage.Set(dto.NewGaugeMetrics(key, value))
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
	return err
}

// SetCounterMetric function sets counter metric to the specified value.
func (s *MetricService) SetCounterMetric(key string, value int64) error {
	s.Logger.Debug(fmt.Sprintf("Key: %s		Value: %d", key, value))

	err := s.storage.Set(dto.NewCounterMetrics(key, value))
	if s.shouldSaveInstantly() {
		s.saveMetrics()
	}
	return err
}

// SetMetrics function updates the list of provided metrics with specified values.
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

// GetMetrics function returns list of requested metric values.
func (s *MetricService) GetMetrics(keys []string) ([]dto.Metrics, error) {
	metrics, err := s.storage.GetMany(keys)
	if err != nil {
		s.Logger.Error(err.Error())
		return nil, err
	}
	return metrics, nil
}

// GetCounterMetric function returns a boolean that indicates existence of the counter metric in database and it's value.
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

// GetGaugeMetric function returns a boolean that indicates existence of the gauge metric in database and it's value.
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

// GetAllMetrics function returns a list of all metrics stored in the database.
func (s *MetricService) GetAllMetrics() ([]dto.Metrics, error) {
	return s.storage.GetAll()
}

// DBHealth function returns a boolean value that indicates accessibility of the database.
func (s *MetricService) DBHealth() bool {
	return s.storage.Health()
}
