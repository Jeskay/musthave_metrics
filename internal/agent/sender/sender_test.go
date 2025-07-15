package sender

import (
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jeskay/musthave_metrics/config"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
)

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
	metricNames := []string{
		"Alloc",
		"HeapIdle",
		"Frees",
		"PollCount",
		"GCCPUFraction",
		"HeapSys",
	}
	reqs := make(chan *http.Request)
	conf := &config.AgentConfig{Address: "localhost:3000"}
	storage := db.NewMemStorage()
	sender := NewHTTPSender(http.DefaultClient, conf, slog.NewTextHandler(os.Stdout, nil), metricNames, []string{})
	storage.Set(dto.NewGaugeMetrics("Alloc", float64(mStats.Alloc)))
	storage.Set(dto.NewGaugeMetrics("HeapIdle", float64(mStats.HeapIdle)))
	storage.Set(dto.NewGaugeMetrics("Frees", float64(mStats.Frees)))
	storage.Set(dto.NewCounterMetrics("PollCount", int64(1)))
	storage.Set(dto.NewGaugeMetrics("GCCPUFraction", mStats.GCCPUFraction))
	storage.Set(dto.NewGaugeMetrics("HeapSys", float64(mStats.HeapSys)))
	metricList, err := storage.GetMany(metricNames)
	if err != nil {
		assert.FailNow(t, "failed due to storage implementation failure")
	}
	go sender.PrepareMetrics(metricList, reqs)
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
