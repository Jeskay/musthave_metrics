package db

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncAccessStorage(t *testing.T) {
	var storage internal.Repositories = NewMemStorage()
	var values sync.Map
	for i := 0; i < 50; i++ {
		valC := dto.NewCounterMetrics(fmt.Sprintf("testCounter%d", i), int64(i))
		valG := dto.NewGaugeMetrics(fmt.Sprintf("testGauge%d", i), float64(i))
		values.Store(valC.ID, valC)
		values.Store(valG.ID, valG)
	}
	var wg sync.WaitGroup
	values.Range(func(key, value any) bool {
		wg.Add(1)
		go func(v dto.Metrics) {
			storage.Set(v)
			wg.Done()
		}(value.(dto.Metrics))
		return true
	})
	wg.Wait()
	require.Len(t, storage.GetAll(), 100)
	res := make(chan bool)
	values.Range(func(key, value any) bool {
		go func(v dto.Metrics, out chan<- bool) {
			value, ok := storage.Get(v.ID)
			if !ok {
				out <- false
				return
			}
			out <- assert.ObjectsAreEqual(v, value)
		}(value.(dto.Metrics), res)
		return true
	})
	for i := 0; i < 100; i++ {
		assert.True(t, <-res)
	}
	close(res)
}

func TestSeqSavingStorage(t *testing.T) {
	var storage internal.Repositories = NewMemStorage()
	valuesC := make([]dto.Metrics, 50)
	valuesG := make([]dto.Metrics, 50)
	for i := 0; i < 50; i++ {
		valC := dto.NewCounterMetrics(fmt.Sprintf("testCounter%d", i), int64(i))
		valG := dto.NewGaugeMetrics(fmt.Sprintf("testGauge%d", i), float64(i))
		valuesC[i] = valC
		valuesG[i] = valG
		storage.Set(valC)
		storage.Set(valG)
	}

	for i := 0; i < 50; i++ {
		valC := valuesC[i]
		valG := valuesG[i]
		counter, ok := storage.Get(fmt.Sprintf("testCounter%d", i))
		assert.True(t, ok)
		gauge, ok := storage.Get(fmt.Sprintf("testGauge%d", i))
		assert.True(t, ok)

		assert.True(
			t,
			assert.ObjectsAreEqual(valG, gauge),
		)
		assert.True(
			t,
			assert.ObjectsAreEqual(valC, counter),
		)

	}
	assert.Len(t, storage.GetAll(), len(valuesC)+len(valuesG))
}

func TestSeqAccessStorage(t *testing.T) {
	var storage internal.Repositories = NewMemStorage()
	obj1 := dto.NewGaugeMetrics("test", float64(9))
	obj2 := dto.NewCounterMetrics("test2", int64(9))
	for i := 0; i < 10; i++ {
		storage.Set(dto.NewCounterMetrics("test", int64(i)))
		storage.Set(obj1)

		storage.Set(dto.NewGaugeMetrics("test2", float64(i)))
		storage.Set(obj2)
	}
	assert.Len(t, storage.GetAll(), 2)

	m, ok := storage.Get("test")
	assert.True(t, ok)
	assert.True(t, assert.ObjectsAreEqual(obj1, m))
	m, ok = storage.Get("test2")
	assert.True(t, ok)
	assert.True(t, assert.ObjectsAreEqual(obj2, m))
	_, ok = storage.Get("unknown")
	assert.False(t, ok)
}
