package agent

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/agent/request"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
	"github.com/Jeskay/musthave_metrics/internal/util"
	"github.com/shirou/gopsutil/mem"
)

type AgentService struct {
	config        *config.AgentConfig
	JsonAvailable bool
	storage       internal.Repositories
	monitorTick   *time.Ticker
	updateTick    *time.Ticker
	pollCount     int64
	serverAddr    string
	logger        *slog.Logger
}

func NewAgentService(conf *config.AgentConfig, logger slog.Handler) *AgentService {
	service := &AgentService{
		storage:    db.NewMemStorage(),
		serverAddr: "http://" + conf.Address,
		logger:     slog.New(logger),
		config:     conf,
	}
	return service
}

func (svc *AgentService) CheckApiAvailability() error {
	var res *http.Response
	err := util.TryRun(func() (err error) {
		res, err = http.Get(svc.serverAddr + "/ping")
		return
	}, util.IsConnectionRefused)

	svc.JsonAvailable = (err == nil) && (res.StatusCode == http.StatusOK)
	return err
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
			reqs := make(chan *http.Request, svc.config.RateLimit)
			if svc.JsonAvailable {
				go svc.PrepareMetricsBatch(metricMainList, reqs, 8)
			} else {
				go svc.PrepareMetrics(metricMainList, reqs)
			}
			go svc.PrepareMetrics(metricSecondaryList, reqs)
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
	svc.storage.Set(dto.NewGaugeMetrics("Alloc", float64(mStats.Alloc)))
	svc.storage.Set(dto.NewGaugeMetrics("BuckHashSys", float64(mStats.BuckHashSys)))
	svc.storage.Set(dto.NewGaugeMetrics("Frees", float64(mStats.Frees)))
	svc.storage.Set(dto.NewGaugeMetrics("GCCPUFraction", float64(mStats.GCCPUFraction)))
	svc.storage.Set(dto.NewGaugeMetrics("GCSys", float64(mStats.GCSys)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapAlloc", float64(mStats.HeapAlloc)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapIdle", float64(mStats.HeapIdle)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapInuse", float64(mStats.HeapInuse)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapObjects", float64(mStats.HeapObjects)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapReleased", float64(mStats.HeapReleased)))
	svc.storage.Set(dto.NewGaugeMetrics("HeapSys", float64(mStats.HeapSys)))
	svc.storage.Set(dto.NewGaugeMetrics("LastGC", float64(mStats.LastGC)))
	svc.storage.Set(dto.NewGaugeMetrics("Lookups", float64(mStats.Lookups)))
	svc.storage.Set(dto.NewGaugeMetrics("MCacheInuse", float64(mStats.MCacheInuse)))
	svc.storage.Set(dto.NewGaugeMetrics("MCacheSys", float64(mStats.MCacheSys)))
	svc.storage.Set(dto.NewGaugeMetrics("MSpanInuse", float64(mStats.MSpanInuse)))
	svc.storage.Set(dto.NewGaugeMetrics("MSpanSys", float64(mStats.MSpanSys)))
	svc.storage.Set(dto.NewGaugeMetrics("Mallocs", float64(mStats.Mallocs)))
	svc.storage.Set(dto.NewGaugeMetrics("NextGC", float64(mStats.NextGC)))
	svc.storage.Set(dto.NewGaugeMetrics("NumForcedGC", float64(mStats.NumForcedGC)))
	svc.storage.Set(dto.NewGaugeMetrics("NumGC", float64(mStats.NumGC)))
	svc.storage.Set(dto.NewGaugeMetrics("OtherSys", float64(mStats.OtherSys)))
	svc.storage.Set(dto.NewGaugeMetrics("PauseTotalNs", float64(mStats.PauseTotalNs)))
	svc.storage.Set(dto.NewGaugeMetrics("StackInuse", float64(mStats.StackInuse)))
	svc.storage.Set(dto.NewGaugeMetrics("StackSys", float64(mStats.StackSys)))
	svc.storage.Set(dto.NewGaugeMetrics("Sys", float64(mStats.Sys)))
	svc.storage.Set(dto.NewGaugeMetrics("TotalAlloc", float64(mStats.TotalAlloc)))

	svc.storage.Set(dto.NewCounterMetrics("PollCount", svc.pollCount))
	rValue := 1e-307 + rand.Float64()*(1e+308-1e-307)
	svc.storage.Set(dto.NewGaugeMetrics("RandomValue", float64(rValue)))

	v, err := mem.VirtualMemory()
	if err == nil {
		svc.storage.Set(dto.NewGaugeMetrics("TotalMemory", float64(v.Total)))
		svc.storage.Set(dto.NewGaugeMetrics("FreeMemory", float64(v.Free)))
		svc.storage.Set(dto.NewGaugeMetrics("CPUutilization1", float64(v.Used)))
	}

}

func (svc *AgentService) PrepareMetrics(metrics []string, requests chan *http.Request) {
	var wg sync.WaitGroup
	for _, metricName := range metrics {
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

func (svc *AgentService) PrepareMetricsBatch(metrics []string, requests chan *http.Request, batchSize int) {
	batch := make([]dto.Metrics, 0)
	url := svc.serverAddr + "/updates/"
	for i, metricName := range metrics {
		metric, ok := svc.storage.Get(metricName)
		if !ok {
			continue
		}
		batch = append(batch, metric)
		if (i+1)%batchSize == 0 {
			r, err := request.MetricsPostJson(svc.config.HashKey, batch, url)
			svc.logger.Debug("post metrics batch", slog.Any("response", r))
			if err != nil {
				svc.logger.Error("batch post response failed", slog.String("error", err.Error()))
				continue
			}
			requests <- r
			batch = make([]dto.Metrics, 0)
		}
	}
	if len(batch) > 0 {
		if r, err := request.MetricsPostJson(svc.config.HashKey, batch, url); err == nil {
			requests <- r
		}
	}
	close(requests)
}

func (svc *AgentService) SendMetrics(requests chan *http.Request) {
	var wg sync.WaitGroup
	for range svc.config.RateLimit {
		wg.Add(1)
		go func() {
			svc.sendWorker(requests)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (svc *AgentService) sendWorker(jobs <-chan *http.Request) {
	for req := range jobs {
		var res *http.Response
		err := util.TryRun(func() (err error) {
			res, err = http.DefaultClient.Do(req)
			return
		}, util.IsConnectionRefused)
		if err != nil {
			svc.logger.Error(err.Error())
		}
		if _, err = io.Copy(io.Discard, res.Body); err != nil {
			svc.logger.Error(err.Error())
		}
		res.Body.Close()

		svc.logger.Debug(fmt.Sprintf("%v", res))
	}
}
