package agent

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectMetrics(t *testing.T) {
	mStats := &runtime.MemStats{
		Alloc:         200,
		BuckHashSys:   100,
		Frees:         1,
		GCCPUFraction: 10,
		PauseNs:       [256]uint64{},
	}
	svc := NewAgentService("localhost", ":3000")
	svc.CollectMetrics(mStats)
	m, _ := svc.storage.Get("Alloc")
	assert.Equal(t, m.Type, internal.GaugeMetric)
	v, ok := m.Value.(uint64)
	require.True(t, ok)
	assert.Equal(t, v, mStats.Alloc)
	val, ok := svc.storage.Get("PauseNs")
	assert.Zero(t, val)
	assert.False(t, ok)

	var i int64
	for i = 1; i < 100; i++ {
		m, ok = svc.storage.Get("PollCount")
		assert.True(t, ok)
		assert.Equal(t, internal.CounterMetric, m.Type)
		poll, ok := m.Value.(int64)
		require.True(t, ok)
		assert.Equal(t, i, poll)
		svc.CollectMetrics(mStats)
	}

	m, ok = svc.storage.Get("RandomValue")
	require.True(t, ok)
	assert.Equal(t, internal.GaugeMetric, m.Type)
	_, ok = m.Value.(float64)
	assert.True(t, ok)
}

func TestPrepareMetrics(t *testing.T) {
	expected := []string{
		"http://localhost:3000/update/gauge/Alloc/200",
		"http://localhost:3000/update/gauge/HeapIdle/0",
		"http://localhost:3000/update/counter/PollCount/1",
		"http://localhost:3000/update/gauge/Frees/1",
		"http://localhost:3000/update/gauge/GCCPUFraction/10.3333",
		"http://localhost:3000/update/gauge/HeapSys/0",
	}
	mStats := &runtime.MemStats{
		Alloc:         200,
		Frees:         1,
		GCCPUFraction: 10.3333,
		HeapSys:       0,
	}
	reqs := make(chan *http.Request)
	svc := NewAgentService("localhost", ":3000")
	svc.storage.Set("Alloc", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Alloc)})
	svc.storage.Set("HeapIdle", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapIdle)})
	svc.storage.Set("Frees", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Frees)})
	svc.storage.Set("PollCount", internal.MetricValue{Type: internal.CounterMetric, Value: int64(1)})
	svc.storage.Set("GCCPUFraction", internal.MetricValue{Type: internal.GaugeMetric, Value: mStats.GCCPUFraction})
	svc.storage.Set("HeapSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapSys)})
	go svc.PrepareMetrics(reqs)
	count := 0
	for r := range reqs {
		assert.Equal(t, http.MethodPost, r.Method)
		req := r.URL.String()
		assert.Contains(t, expected, req)
		count++
	}
	assert.Equal(t, len(expected), count)
}
