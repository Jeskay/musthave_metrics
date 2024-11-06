package db

import (
	"sync"

	"github.com/Jeskay/musthave_metrics/internal"
)

type MemStorage struct {
	data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: sync.Map{},
	}
}

func (ms *MemStorage) Set(key string, value internal.MetricValue) {
	ms.data.Store(key, value)
}

func (ms *MemStorage) Get(key string) (internal.MetricValue, bool) {
	if m, ok := ms.data.Load(key); ok {
		return m.(internal.MetricValue), ok
	}
	return internal.MetricValue{}, false
}

func (ms *MemStorage) GetAll() []*internal.Metric {
	m := make([]*internal.Metric, 0)
	ms.data.Range(func(key, value any) bool {
		m = append(m, &internal.Metric{Name: key.(string), Value: value.(internal.MetricValue)})
		return true
	})
	return m
}

func (ms *MemStorage) Health() bool { return true }
