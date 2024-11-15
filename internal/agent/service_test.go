package agent

import (
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
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
	svc := NewAgentService("localhost:3000", slog.NewTextHandler(os.Stdout, nil))
	svc.CollectMetrics(mStats)
	m, _ := svc.storage.Get("Alloc")
	assert.Equal(t, string(internal.GaugeMetric), m.MType)
	require.True(t, m.Value != nil)
	assert.Equal(t, *m.Value, float64(mStats.Alloc))
	val, ok := svc.storage.Get("PauseNs")
	assert.Zero(t, val)
	assert.False(t, ok)

	var i int64
	var sum int64 = 0
	for i = 1; i < 100; i++ {
		sum += i
		m, ok = svc.storage.Get("PollCount")
		assert.True(t, ok)
		assert.Equal(t, string(internal.CounterMetric), m.MType)
		require.True(t, m.Delta != nil)
		assert.Equal(t, sum, *m.Delta)
		svc.CollectMetrics(mStats)
	}

	m, ok = svc.storage.Get("RandomValue")
	require.True(t, ok)
	assert.Equal(t, string(internal.GaugeMetric), m.MType)
	assert.True(t, m.Value != nil)
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
	svc := NewAgentService("localhost:3000", slog.NewTextHandler(os.Stdout, nil))
	svc.storage.Set(dto.NewGaugeMetrics("Alloc", float64(mStats.Alloc)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapIdle", float64(mStats.HeapIdle)))
	svc.storage.Set(dto.NewGaugeMetrics("Frees", float64(mStats.Frees)))
	svc.storage.Set(dto.NewCounterMetrics("PollCount", int64(1)))
	svc.storage.Set(dto.NewGaugeMetrics("GCCPUFraction", mStats.GCCPUFraction))
	svc.storage.Set(dto.NewGaugeMetrics("HeapSys", float64(mStats.HeapSys)))
	go svc.PrepareMetrics(reqs)
	count := 0
	jsonCount := 0
	for r := range reqs {
		assert.Equal(t, http.MethodPost, r.Method)
		req := r.URL.String()
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			assert.Equal(t, "http://localhost:3000/update/", req)
			jsonCount++
		} else {
			assert.Contains(t, expected, req)
			count++
		}
	}
	assert.Equal(t, len(expected), jsonCount)
	assert.Equal(t, len(expected), count)
}
