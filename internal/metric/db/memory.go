package db

import (
	"fmt"
	"sync"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

type MemStorage struct {
	data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: sync.Map{},
	}
}

func (ms *MemStorage) Set(value dto.Metrics) error {
	ms.data.Store(value.ID, value)
	return nil
}

func (ms *MemStorage) SetMany(values []dto.Metrics) error {
	for _, v := range values {
		ms.data.Store(v.ID, v)
	}
	return nil
}

func (ms *MemStorage) Get(key string) (dto.Metrics, bool) {
	if m, ok := ms.data.Load(key); ok {
		return m.(dto.Metrics), ok
	}
	return dto.Metrics{}, false
}

func (ms *MemStorage) GetMany(keys []string) ([]dto.Metrics, error) {
	m := make([]dto.Metrics, len(keys))
	for _, key := range keys {
		if value, ok := ms.Get(key); ok {
			m = append(m, value)
		} else {
			return nil, fmt.Errorf("key %s does not exists", key)
		}
	}
	return m, nil
}

func (ms *MemStorage) GetAll() []dto.Metrics {
	m := make([]dto.Metrics, 0)
	ms.data.Range(func(key, value any) bool {
		v, ok := value.(dto.Metrics)
		if ok {
			m = append(m, v)
		}
		return ok
	})
	return m
}

func (ms *MemStorage) Health() bool { return true }
