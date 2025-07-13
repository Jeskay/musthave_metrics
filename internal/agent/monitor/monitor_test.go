package monitor

import (
	"log/slog"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
)

func TestCollectMetrics(t *testing.T) {
	mStats := &runtime.MemStats{
		Alloc:         200,
		BuckHashSys:   100,
		Frees:         1,
		GCCPUFraction: 10,
		PauseNs:       [256]uint64{},
	}
	monitorMetrics := NewMetricMonitor(slog.NewTextHandler(os.Stdout, nil))
	storage := db.NewMemStorage()
	monitorMetrics.CollectMetrics(mStats, storage.SetMany)
	m, _ := storage.Get("Alloc")
	assert.Equal(t, string(internal.GaugeMetric), m.MType)
	require.True(t, m.Value != nil)
	assert.Equal(t, *m.Value, float64(mStats.Alloc))
	val, ok := storage.Get("PauseNs")
	assert.Zero(t, val)
	assert.False(t, ok)

	var i int64
	var sum int64 = 0
	for i = 1; i < 100; i++ {
		sum += i
		m, ok = storage.Get("PollCount")
		assert.True(t, ok)
		assert.Equal(t, string(internal.CounterMetric), m.MType)
		require.True(t, m.Delta != nil)
		assert.Equal(t, sum, *m.Delta)
		monitorMetrics.CollectMetrics(mStats, storage.SetMany)
	}

	m, ok = storage.Get("RandomValue")
	require.True(t, ok)
	assert.Equal(t, string(internal.GaugeMetric), m.MType)
	assert.True(t, m.Value != nil)
}
