package internal

import "sync"

type MemStorage struct {
	sync.RWMutex
	data map[string]MetricValue
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]MetricValue),
	}
}

func (ms *MemStorage) Set(key string, value MetricValue) {
	ms.Lock()
	ms.data[key] = value
	ms.Unlock()
}

func (ms *MemStorage) Get(key string) (MetricValue, bool) {
	ms.RLock()
	m, ok := ms.data[key]
	ms.RUnlock()
	return m, ok
}

func (ms *MemStorage) GetAll() []*Metric {
	m := make([]*Metric, len(ms.data))
	ms.RLock()
	counter := 0
	for name, metric := range ms.data {
		m[counter] = &Metric{Name: name, Value: metric}
		counter++
	}
	ms.RUnlock()
	return m
}
