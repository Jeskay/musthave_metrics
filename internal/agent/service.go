package agent

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/Jeskay/musthave_metrics/internal"
)

type AgentService struct {
	storage     internal.Repositories
	monitorTick *time.Ticker
	updateTick  *time.Ticker
	pollCount   int64
	serverAddr  string
}

func NewAgentService(host string, port string) *AgentService {
	service := &AgentService{
		storage:    internal.NewMemStorage(),
		serverAddr: "http://" + host + port,
	}
	return service
}

func (svc *AgentService) StartMonitoring(interval time.Duration) chan<- bool {
	if svc.monitorTick != nil {
		return nil
	}
	svc.monitorTick = time.NewTicker(interval)
	quit := make(chan bool)
	go func() {
		mStats := &runtime.MemStats{}
		for {
			select {
			case <-quit:
				svc.monitorTick.Stop()
				return
			case t := <-svc.monitorTick.C:
				runtime.ReadMemStats(mStats)
				svc.CollectMetrics(mStats)
				fmt.Println("Tick at ", t)
			}
		}
	}()

	return quit
}

func (svc *AgentService) StartSending(interval time.Duration) chan<- bool {
	if svc.updateTick != nil {
		return nil
	}
	quit := make(chan bool)
	reqs := make(chan *http.Request)
	svc.updateTick = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-quit:
				svc.updateTick.Stop()
				return
			case t := <-svc.monitorTick.C:
				fmt.Println("Tick at ", t)
				go svc.PrepareMetrics(reqs)
				svc.SendMetrics(reqs)
			}
		}
	}()
	return quit
}

func (svc *AgentService) CollectMetrics(mStats *runtime.MemStats) {
	svc.pollCount++
	svc.storage.Set("Alloc", internal.Metric{Type: internal.GaugeMetric, Value: mStats.Alloc})
	svc.storage.Set("BuckHashSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.BuckHashSys})
	svc.storage.Set("Frees", internal.Metric{Type: internal.GaugeMetric, Value: mStats.Frees})
	svc.storage.Set("GCCPUFraction", internal.Metric{Type: internal.GaugeMetric, Value: mStats.GCCPUFraction})
	svc.storage.Set("GCSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.GCSys})
	svc.storage.Set("HeapAlloc", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapAlloc})
	svc.storage.Set("HeapIdle", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapIdle})
	svc.storage.Set("HeapInuse", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapInuse})
	svc.storage.Set("HeapObjects", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapObjects})
	svc.storage.Set("HeapReleased", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapReleased})
	svc.storage.Set("HeapSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.HeapSys})
	svc.storage.Set("LastGC", internal.Metric{Type: internal.GaugeMetric, Value: mStats.LastGC})
	svc.storage.Set("Lookups", internal.Metric{Type: internal.GaugeMetric, Value: mStats.Lookups})
	svc.storage.Set("MCacheInuse", internal.Metric{Type: internal.GaugeMetric, Value: mStats.MCacheInuse})
	svc.storage.Set("MCacheSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.MCacheSys})
	svc.storage.Set("MSpanInuse", internal.Metric{Type: internal.GaugeMetric, Value: mStats.MSpanInuse})
	svc.storage.Set("MSpanSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.MSpanSys})
	svc.storage.Set("Mallocs", internal.Metric{Type: internal.GaugeMetric, Value: mStats.Mallocs})
	svc.storage.Set("NextGC", internal.Metric{Type: internal.GaugeMetric, Value: mStats.NextGC})
	svc.storage.Set("NumForcedGC", internal.Metric{Type: internal.GaugeMetric, Value: mStats.NumForcedGC})
	svc.storage.Set("NumGC", internal.Metric{Type: internal.GaugeMetric, Value: mStats.NumGC})
	svc.storage.Set("OtherSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.OtherSys})
	svc.storage.Set("PauseTotalNs", internal.Metric{Type: internal.GaugeMetric, Value: mStats.PauseTotalNs})
	svc.storage.Set("StackInuse", internal.Metric{Type: internal.GaugeMetric, Value: mStats.StackInuse})
	svc.storage.Set("StackSys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.StackSys})
	svc.storage.Set("Sys", internal.Metric{Type: internal.GaugeMetric, Value: mStats.Sys})
	svc.storage.Set("TotalAlloc", internal.Metric{Type: internal.GaugeMetric, Value: mStats.TotalAlloc})

	svc.storage.Set("PollCount", internal.Metric{Type: internal.CounterMetric, Value: svc.pollCount})
	rValue := math.SmallestNonzeroFloat64 + rand.Float64()*(math.MaxFloat64-math.SmallestNonzeroFloat64)
	svc.storage.Set("RandomValue", internal.Metric{Type: internal.GaugeMetric, Value: rValue})

}

func (svc *AgentService) PrepareMetrics(requests chan *http.Request) {
	for _, metricName := range metricList {
		metric, ok := svc.storage.Get(metricName)
		if !ok {
			continue
		}
		url := svc.serverAddr + "/update/"
		if metric.Type == internal.CounterMetric {
			v, ok := metric.Value.(int64)
			if !ok {
				v = 0
			}
			url += path.Join(string(metric.Type), metricName, strconv.FormatInt(v, 10))
		} else if metric.Type == internal.GaugeMetric {
			v, ok := metric.Value.(float64)
			if !ok {
				v = 0
			}
			url += path.Join(string(metric.Type), metricName, strconv.FormatFloat(v, 'f', -1, 64))
		}
		r, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		requests <- r
	}
	close(requests)
}

func (svc *AgentService) SendMetrics(requests chan *http.Request) {
	for req := range requests {
		go func(req *http.Request) {
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Error", err)
			}
			fmt.Println(r)
		}(req)
	}
}
