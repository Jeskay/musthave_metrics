package internal

import "sync"

type MemStorage struct {
	data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: sync.Map{},
	}
}

func (ms *MemStorage) Set(key string, value MetricValue) {
	ms.data.Store(key, value)
}

func (ms *MemStorage) Get(key string) (MetricValue, bool) {
	if m, ok := ms.data.Load(key); ok {
		return m.(MetricValue), ok
	}
	return MetricValue{}, false
}

func (ms *MemStorage) GetAll() []*Metric {
	m := make([]*Metric, 0)
	ms.data.Range(func(key, value any) bool {
		m = append(m, &Metric{Name: key.(string), Value: value.(MetricValue)})
		return true
	})
	return m
}
