package monitor

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"runtime"
	"time"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/shirou/gopsutil/mem"
)

type MetricMonitor struct {
	monitorTick *time.Ticker
	pollCount   int64
	logger      *slog.Logger
}

func NewMetricMonitor(logger slog.Handler) *MetricMonitor {
	return &MetricMonitor{
		pollCount: 0,
		logger:    slog.New(logger),
	}
}

// StartMonitoring function initiates the process of collecting memory metrics.
func (m *MetricMonitor) StartMonitoring(interval time.Duration, save func([]dto.Metrics) error) chan<- struct{} {
	if m.monitorTick != nil {
		return nil
	}
	m.monitorTick = time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		mStats := &runtime.MemStats{}
	loop:
		for {
			runtime.ReadMemStats(mStats)
			m.CollectMetrics(mStats, save)
			select {
			case t := <-m.monitorTick.C:
				m.logger.Debug(fmt.Sprintf("Tick at %s", t.String()))
				continue
			case <-quit:
				m.monitorTick.Stop()
				break loop
			}
		}
	}()

	return quit
}

// CollectMetrics function saves current memory data to the provided storage.
func (m *MetricMonitor) CollectMetrics(mStats *runtime.MemStats, save func([]dto.Metrics) error) {
	m.pollCount++
	rValue := 1e-307 + rand.Float64()*(1e+308-1e-307)
	err := save([]dto.Metrics{
		dto.NewGaugeMetrics("Alloc", float64(mStats.Alloc)),
		dto.NewGaugeMetrics("BuckHashSys", float64(mStats.BuckHashSys)),
		dto.NewGaugeMetrics("Frees", float64(mStats.Frees)),
		dto.NewGaugeMetrics("GCCPUFraction", float64(mStats.GCCPUFraction)),
		dto.NewGaugeMetrics("GCSys", float64(mStats.GCSys)),
		dto.NewGaugeMetrics("HeapAlloc", float64(mStats.HeapAlloc)),
		dto.NewGaugeMetrics("HeapIdle", float64(mStats.HeapIdle)),
		dto.NewGaugeMetrics("HeapInuse", float64(mStats.HeapInuse)),
		dto.NewGaugeMetrics("HeapObjects", float64(mStats.HeapObjects)),
		dto.NewGaugeMetrics("HeapReleased", float64(mStats.HeapReleased)),
		dto.NewGaugeMetrics("HeapSys", float64(mStats.HeapSys)),
		dto.NewGaugeMetrics("LastGC", float64(mStats.LastGC)),
		dto.NewGaugeMetrics("Lookups", float64(mStats.Lookups)),
		dto.NewGaugeMetrics("MCacheInuse", float64(mStats.MCacheInuse)),
		dto.NewGaugeMetrics("MCacheSys", float64(mStats.MCacheSys)),
		dto.NewGaugeMetrics("MSpanInuse", float64(mStats.MSpanInuse)),
		dto.NewGaugeMetrics("MSpanSys", float64(mStats.MSpanSys)),
		dto.NewGaugeMetrics("Mallocs", float64(mStats.Mallocs)),
		dto.NewGaugeMetrics("NextGC", float64(mStats.NextGC)),
		dto.NewGaugeMetrics("NumForcedGC", float64(mStats.NumForcedGC)),
		dto.NewGaugeMetrics("NumGC", float64(mStats.NumGC)),
		dto.NewGaugeMetrics("OtherSys", float64(mStats.OtherSys)),
		dto.NewGaugeMetrics("PauseTotalNs", float64(mStats.PauseTotalNs)),
		dto.NewGaugeMetrics("StackInuse", float64(mStats.StackInuse)),
		dto.NewGaugeMetrics("StackSys", float64(mStats.StackSys)),
		dto.NewGaugeMetrics("Sys", float64(mStats.Sys)),
		dto.NewGaugeMetrics("TotalAlloc", float64(mStats.TotalAlloc)),
		dto.NewCounterMetrics("PollCount", m.pollCount),
		dto.NewGaugeMetrics("RandomValue", float64(rValue)),
	})
	if err != nil {
		m.logger.Error(err.Error())
	}

	if v, err := mem.VirtualMemory(); err == nil {
		err := save([]dto.Metrics{
			dto.NewGaugeMetrics("TotalMemory", float64(v.Total)),
			dto.NewGaugeMetrics("FreeMemory", float64(v.Free)),
			dto.NewGaugeMetrics("CPUutilization1", float64(v.Used)),
		})
		if err != nil {
			m.logger.Error(err.Error())
		}
	} else {
		m.logger.Error(err.Error())
	}
}
