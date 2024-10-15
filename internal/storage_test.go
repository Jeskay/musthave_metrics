package internal

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncAccessStorage(t *testing.T) {
	var storage Repositories = NewMemStorage()
	var values sync.Map
	for i := 0; i < 50; i++ {
		valC := MetricValue{Type: CounterMetric, Value: int64(i)}
		valG := MetricValue{Type: GaugeMetric, Value: float64(i)}
		values.Store(fmt.Sprintf("testCounter%d", i), valC)
		values.Store(fmt.Sprintf("testGauge%d", i), valG)
	}
	var wg sync.WaitGroup
	values.Range(func(key, value any) bool {
		wg.Add(1)
		go func(k string, v MetricValue) {
			storage.Set(k, v)
			wg.Done()
		}(key.(string), value.(MetricValue))
		return true
	})
	wg.Wait()
	require.Len(t, storage.GetAll(), 100)
	res := make(chan bool)
	values.Range(func(key, value any) bool {
		go func(k string, v MetricValue, out chan<- bool) {
			value, ok := storage.Get(k)
			if !ok {
				out <- false
				return
			}
			out <- assert.ObjectsAreEqual(v, value)
		}(key.(string), value.(MetricValue), res)
		return true
	})
	for i := 0; i < 100; i++ {
		assert.True(t, <-res)
	}
	close(res)
}

func TestSeqSavingStorage(t *testing.T) {
	var storage Repositories = NewMemStorage()
	valuesC := make([]MetricValue, 50)
	valuesG := make([]MetricValue, 50)
	for i := 0; i < 50; i++ {
		valC := MetricValue{Type: CounterMetric, Value: int64(i)}
		valG := MetricValue{Type: GaugeMetric, Value: float64(i)}
		valuesC[i] = valC
		valuesG[i] = valG
		storage.Set(fmt.Sprintf("testCounter%d", i), valC)
		storage.Set(fmt.Sprintf("testGauge%d", i), valG)
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
	var storage Repositories = NewMemStorage()
	obj1 := MetricValue{Type: GaugeMetric, Value: float64(9)}
	obj2 := MetricValue{Type: CounterMetric, Value: int64(9)}
	for i := 0; i < 10; i++ {
		storage.Set("test", MetricValue{Type: CounterMetric, Value: int64(i)})
		storage.Set("test", obj1)

		storage.Set("test2", MetricValue{Type: GaugeMetric, Value: float64(i)})
		storage.Set("test2", obj2)
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
