package internal

import "sync"

type MemStorage struct {
	sync.RWMutex
	data map[string]Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]Metric),
	}
}

func (ms *MemStorage) Set(key string, value Metric) {
	ms.Lock()
	ms.data[key] = value
	ms.Unlock()
}

func (ms *MemStorage) Get(key string) (Metric, bool) {
	ms.RLock()
	m, ok := ms.data[key]
	ms.RUnlock()
	return m, ok
}
