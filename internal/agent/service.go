// Package agent contains the business logic and data structures of metric agent.
// The agent sends various memory data over small intervals to the server which collects and stores it.
package agent

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/agent/request"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
	"github.com/Jeskay/musthave_metrics/internal/util"
	"github.com/Jeskay/musthave_metrics/pkg/worker"
	pb "github.com/Jeskay/musthave_metrics/protos"
)

// AgentService struct provides the functionality of collecting and sending metric data to the server.
type AgentService struct {
	workerPoolHTTP *worker.WorkerPool[*http.Request]
	workerPoolGRPC *worker.WorkerPool[*pb.Metric]
	clientHTTP     *http.Client
	clientGRPC     pb.ServerClient
	cipherService  *request.Cipher
	config         *config.AgentConfig
	ipAddress      net.IP
	JsonAvailable  bool
	storage        internal.Repositories
	monitorTick    *time.Ticker
	updateTick     *time.Ticker
	pollCount      int64
	serverAddr     string
	logger         *slog.Logger
}

// NewAgentService function initializes and returns new instance of AgentService.
func NewAgentService(client *http.Client, conf *config.AgentConfig, logger slog.Handler) *AgentService {
	service := &AgentService{
		clientHTTP:     client,
		storage:        db.NewMemStorage(),
		serverAddr:     "http://" + conf.Address,
		logger:         slog.New(logger),
		config:         conf,
		workerPoolHTTP: worker.NewWorkerPool[*http.Request](conf.RateLimit),
		workerPoolGRPC: worker.NewWorkerPool[*pb.Metric](conf.RateLimit),
	}
	if conf.GRPC {
		conn, err := grpc.NewClient(conf.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			service.logger.Error("cannot establish grpc connection", err)
		}
		service.clientGRPC = pb.NewServerClient(conn)
	}
	cipherService, err := request.NewCipher(conf.PublicKey)
	if err != nil {
		service.logger.Error("failed to initialize cipher service")
		cipherService = nil
	}
	ip, err := util.GetSelfIP()
	if err != nil {
		service.logger.Error("unknown issue when establishing logical connection to DNS", err)
	}

	service.cipherService = cipherService
	service.ipAddress = ip
	return service
}

// CheckAPIAvailability of the metric server and returns error if it is unaccessible.
func (svc *AgentService) CheckAPIAvailability() error {
	res, err := http.Get(svc.serverAddr + "/ping")
	if res != nil {
		defer res.Body.Close()
	}
	svc.JsonAvailable = (err == nil) && (res.StatusCode == http.StatusOK)
	return err
}

// StartMonitoring function initiates the process of collecting memory metrics to store in agent's memory storage.
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

// StartSending function initiates the process of sending collected data from in-memory storage to the metric server.
// If JSON format is supported by the API, the agent will send metrics in batches thus saving on the amount of requests.
func (svc *AgentService) StartSending(interval time.Duration) chan<- struct{} {
	if svc.updateTick != nil {
		return nil
	}
	quit := make(chan struct{})
	var finishWg sync.WaitGroup
	svc.updateTick = time.NewTicker(interval)
	go func() {

	loop:
		for {
			finishWg.Add(1)
			if svc.config.GRPC {
				reqs := make(chan *pb.Metric, svc.config.RateLimit)
				go func() {
					svc.PrepareMetricsGRPC(metricMainList, reqs)
					svc.PrepareMetricsGRPC(metricSecondaryList, reqs)
					close(reqs)
					finishWg.Done()
				}()
				go svc.SendMetricsGRPC(reqs)
			} else {
				reqs := make(chan *http.Request, svc.config.RateLimit)
				go func() {
					if svc.JsonAvailable {
						svc.PrepareMetricsBatch(metricMainList, reqs, 8)
					} else {
						svc.PrepareMetrics(metricMainList, reqs)
					}
					svc.PrepareMetrics(metricSecondaryList, reqs)
					close(reqs)
					finishWg.Done()
				}()
				go svc.SendMetrics(reqs)
			}

			select {
			case t := <-svc.monitorTick.C:
				svc.logger.Debug(fmt.Sprintf("Tick at %s", t.String()))
				continue
			case <-quit:
				svc.updateTick.Stop()
				finishWg.Wait()
				break loop

			}
		}
	}()
	return quit
}

// CollectMetrics function saves current memory data to the memory storage of the agent.
func (svc *AgentService) CollectMetrics(mStats *runtime.MemStats) {
	svc.pollCount++
	rValue := 1e-307 + rand.Float64()*(1e+308-1e-307)
	err := svc.storage.SetMany([]dto.Metrics{
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
		dto.NewCounterMetrics("PollCount", svc.pollCount),
		dto.NewGaugeMetrics("RandomValue", float64(rValue)),
	})
	if err != nil {
		svc.logger.Error(err.Error())
	}

	if v, err := mem.VirtualMemory(); err == nil {
		err := svc.storage.SetMany([]dto.Metrics{
			dto.NewGaugeMetrics("TotalMemory", float64(v.Total)),
			dto.NewGaugeMetrics("FreeMemory", float64(v.Free)),
			dto.NewGaugeMetrics("CPUutilization1", float64(v.Used)),
		})
		if err != nil {
			svc.logger.Error(err.Error())
		}
	} else {
		svc.logger.Error(err.Error())
	}
}

// PrepareMetrics function assembles metrics data from agent's storage into HTTP requests to send.
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
			rj, err := request.MetricPostJson(svc.ipAddress, svc.config.HashKey, svc.cipherService, metric, url)
			if err != nil {
				svc.logger.Error(err.Error())
				return
			}
			requests <- rt
			requests <- rj
		}()
	}
	wg.Wait()
}

func (svc *AgentService) PrepareMetricsGRPC(metricNames []string, metrics chan *pb.Metric) {
	var wg sync.WaitGroup
	for _, metricName := range metricNames {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metric, ok := svc.storage.Get(metricName)
			if !ok {
				return
			}
			metrics <- &pb.Metric{Id: metric.ID, Type: metric.MType, Delta: metric.Delta, Value: metric.Value}
		}()
	}
	wg.Wait()
}

// PrepareMetricsBatch function assembles metrics data from agent's storage
// into batch HTTP requests of specified size to send to the server.
func (svc *AgentService) PrepareMetricsBatch(metrics []string, requests chan *http.Request, batchSize int) {
	batch := make([]dto.Metrics, 0)
	url := svc.serverAddr + "/updates"
	for i, metricName := range metrics {
		metric, ok := svc.storage.Get(metricName)
		if !ok {
			continue
		}
		batch = append(batch, metric)
		if (i+1)%batchSize == 0 {
			r, err := request.MetricsPostJson(svc.ipAddress, svc.config.HashKey, svc.cipherService, batch, url)
			if err != nil {
				svc.logger.Error("batch post response failed", slog.String("error", err.Error()))
				continue
			}
			svc.logger.Debug("post metrics batch", slog.Any("response", r))
			requests <- r
			batch = make([]dto.Metrics, 0)
		}
	}
	if len(batch) > 0 {
		if r, err := request.MetricsPostJson(svc.ipAddress, svc.config.HashKey, svc.cipherService, batch, url); err == nil {
			requests <- r
		}
	}
}

// SendMetrics function starts sending of the prepared HTTP requests to metric server.
func (svc *AgentService) SendMetrics(requests chan *http.Request) {
	svc.workerPoolHTTP.Run(requests, func(req *http.Request) {
		err := util.TryRun(func() (err error) {
			res, err := svc.clientHTTP.Do(req)
			if res != nil {
				defer res.Body.Close()
			}
			if err != nil {
				return
			}
			if _, err = io.Copy(io.Discard, res.Body); err != nil {
				return
			}
			return
		}, util.IsConnectionRefused)
		if err != nil {
			svc.logger.Error(err.Error())
		}
	})
}

func (svc *AgentService) SendMetricsGRPC(metrics chan *pb.Metric) {
	svc.workerPoolGRPC.Run(metrics, func(m *pb.Metric) {
		err := util.TryRun(func() (err error) {
			ctx := metadata.AppendToOutgoingContext(context.Background(), "uri", svc.ipAddress.String())
			res, err := svc.clientGRPC.UpdateMetric(ctx, &pb.UpdateMetricRequest{Metric: m})
			if err != nil {
				return
			}
			if res != nil {
				svc.logger.Info("Response", slog.String("Value", res.Value.String()))
			}
			return
		}, util.IsConnectionRefused)
		if err != nil {
			svc.logger.Error(err.Error())
		}
	})
}
