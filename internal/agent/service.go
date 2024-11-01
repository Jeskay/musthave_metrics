package agent

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/Jeskay/musthave_metrics/internal/agent/request"
)

type AgentService struct {
	storage     internal.Repositories
	monitorTick *time.Ticker
	updateTick  *time.Ticker
	pollCount   int64
	serverAddr  string
	logger      *slog.Logger
}

func NewAgentService(address string, logger slog.Handler) *AgentService {
	service := &AgentService{
		storage:    internal.NewMemStorage(),
		serverAddr: "http://" + address,
		logger:     slog.New(logger),
	}
	return service
}

func (svc *AgentService) StartMonitoring(interval time.Duration) chan<- struct{} {
	if svc.monitorTick != nil {
		return nil
	}
	svc.monitorTick = time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		mStats := &runtime.MemStats{}
	loop:
		for {
			runtime.ReadMemStats(mStats)
			svc.CollectMetrics(mStats)
			select {
			case t := <-svc.monitorTick.C:
				svc.logger.Debug(fmt.Sprintf("Tick at %s", t.String()))
				continue
			case <-quit:
				svc.monitorTick.Stop()
				break loop
			}
		}
	}()

	return quit
}

func (svc *AgentService) StartSending(interval time.Duration) chan<- struct{} {
	if svc.updateTick != nil {
		return nil
	}
	quit := make(chan struct{})
	svc.updateTick = time.NewTicker(interval)
	go func() {

	loop:
		for {
			reqs := make(chan *http.Request)
			go svc.PrepareMetrics(reqs)
			go svc.SendMetrics(reqs)
			select {
			case t := <-svc.monitorTick.C:
				svc.logger.Debug(fmt.Sprintf("Tick at %s", t.String()))
				continue
			case <-quit:
				svc.updateTick.Stop()
				break loop

			}
		}
	}()
	return quit
}

func (svc *AgentService) CollectMetrics(mStats *runtime.MemStats) {
	svc.pollCount++
	svc.storage.Set("Alloc", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Alloc)})
	svc.storage.Set("BuckHashSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.BuckHashSys)})
	svc.storage.Set("Frees", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Frees)})
	svc.storage.Set("GCCPUFraction", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.GCCPUFraction)})
	svc.storage.Set("GCSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.GCSys)})
	svc.storage.Set("HeapAlloc", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapAlloc)})
	svc.storage.Set("HeapIdle", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapIdle)})
	svc.storage.Set("HeapInuse", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapInuse)})
	svc.storage.Set("HeapObjects", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapObjects)})
	svc.storage.Set("HeapReleased", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapReleased)})
	svc.storage.Set("HeapSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.HeapSys)})
	svc.storage.Set("LastGC", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.LastGC)})
	svc.storage.Set("Lookups", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Lookups)})
	svc.storage.Set("MCacheInuse", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.MCacheInuse)})
	svc.storage.Set("MCacheSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.MCacheSys)})
	svc.storage.Set("MSpanInuse", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.MSpanInuse)})
	svc.storage.Set("MSpanSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.MSpanSys)})
	svc.storage.Set("Mallocs", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Mallocs)})
	svc.storage.Set("NextGC", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.NextGC)})
	svc.storage.Set("NumForcedGC", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.NumForcedGC)})
	svc.storage.Set("NumGC", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.NumGC)})
	svc.storage.Set("OtherSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.OtherSys)})
	svc.storage.Set("PauseTotalNs", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.PauseTotalNs)})
	svc.storage.Set("StackInuse", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.StackInuse)})
	svc.storage.Set("StackSys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.StackSys)})
	svc.storage.Set("Sys", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.Sys)})
	svc.storage.Set("TotalAlloc", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(mStats.TotalAlloc)})

	svc.storage.Set("PollCount", internal.MetricValue{Type: internal.CounterMetric, Value: svc.pollCount})
	rValue := math.SmallestNonzeroFloat64 + rand.Float64()*(math.MaxFloat64-math.SmallestNonzeroFloat64)
	svc.storage.Set("RandomValue", internal.MetricValue{Type: internal.GaugeMetric, Value: float64(rValue)})

}

func (svc *AgentService) PrepareMetrics(requests chan *http.Request) {
	var wg sync.WaitGroup
	for _, metricName := range metricList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metric, ok := svc.storage.Get(metricName)
			if !ok {
				return
			}
			url := svc.serverAddr + "/update/"
			rt, err := request.MetricPostPlain(metricName, metric, url)
			if err != nil {
				svc.logger.Error(err.Error())
				return
			}
			rj, err := request.MetricPostJson(metricName, metric, url)
			if err != nil {
				svc.logger.Error(err.Error())
				return
			}
			requests <- rt
			requests <- rj
		}()
	}
	wg.Wait()
	close(requests)
}

func (svc *AgentService) SendMetrics(requests chan *http.Request) {
	var wg sync.WaitGroup
	for req := range requests {
		wg.Add(1)
		go func(req *http.Request) {
			defer wg.Done()
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				svc.logger.Error(err.Error())
				return
			}

			if _, err = io.Copy(io.Discard, res.Body); err != nil {
				svc.logger.Error(err.Error())
			}
			res.Body.Close()

			svc.logger.Debug(fmt.Sprintf("%v", res))
		}(req)
	}
	wg.Wait()
}
