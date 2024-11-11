package db

import (
	"fmt"
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

func (ms *MemStorage) Set(key string, value internal.MetricValue) error {
	ms.data.Store(key, value)
	return nil
}

func (ms *MemStorage) SetMany(values []internal.Metric) error {
	for _, v := range values {
		ms.data.Store(v.Name, v.Value)
	}
	return nil
}

func (ms *MemStorage) Get(key string) (internal.MetricValue, bool) {
	if m, ok := ms.data.Load(key); ok {
		return m.(internal.MetricValue), ok
	}
	return internal.MetricValue{}, false
}

func (ms *MemStorage) GetMany(keys []string) ([]*internal.Metric, error) {
	m := make([]*internal.Metric, len(keys))
	for _, key := range keys {
		if value, ok := ms.Get(key); ok {
			m = append(m, &internal.Metric{Name: key, Value: value})
		} else {
			return nil, fmt.Errorf("key %s does not exists", key)
		}
	}
	return m, nil
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
