package sender

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/agent/cipher"
	"github.com/Jeskay/musthave_metrics/internal/util"
	"github.com/Jeskay/musthave_metrics/pkg/worker"
)

type HTTPSender struct {
	updateTick       *time.Ticker
	workerPool       *worker.WorkerPool[*http.Request]
	client           *http.Client
	addressServer    string
	addressClient    net.IP
	logger           *slog.Logger
	cipherService    *cipher.Cipher
	config           *config.AgentConfig
	JsonAvailable    bool
	primaryMetrics   []string
	secondaryMetrics []string
}

func NewHTTPSender(client *http.Client, conf *config.AgentConfig, logger slog.Handler, primaryMetrics []string, secondaryMetrics []string) *HTTPSender {
	sender := &HTTPSender{
		client:           client,
		addressServer:    "http://" + conf.Address,
		logger:           slog.New(logger),
		config:           conf,
		workerPool:       worker.NewWorkerPool[*http.Request](conf.RateLimit),
		primaryMetrics:   primaryMetrics,
		secondaryMetrics: secondaryMetrics,
	}
	cipherService, err := cipher.NewCipher(conf.PublicKey)
	if err != nil {
		sender.logger.Error("failed to initialize cipher service")
	}
	ip, err := util.GetSelfIP()
	if err != nil {
		sender.logger.Error("unknown issue when establishing logical connection to DNS", slog.Attr{Key: "error", Value: slog.AnyValue(err)})
	}
	sender.cipherService = cipherService
	sender.addressClient = ip
	return sender
}

// PrepareMetrics function assembles metrics data from agent's storage into HTTP requests to send.
func (s *HTTPSender) PrepareMetrics(metrics []dto.Metrics, requests chan *http.Request) {
	var wg sync.WaitGroup
	for _, metric := range metrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url := s.addressServer + "/update/"
			rt, err := MetricPostPlain(metric.ID, metric, url)
			if err != nil {
				s.logger.Error(err.Error())
				return
			}
			rj, err := MetricPostJson(s.addressClient, s.config.HashKey, s.cipherService, metric, url)
			if err != nil {
				s.logger.Error(err.Error())
				return
			}
			requests <- rt
			requests <- rj
		}()
	}
	wg.Wait()
}

// CheckAPIAvailability of the metric server and returns error if it is unaccessible.
func (s *HTTPSender) CheckAPIAvailability() error {
	res, err := http.Get(s.addressServer + "/ping")
	if res != nil {
		defer res.Body.Close()
	}
	s.JsonAvailable = (err == nil) && (res.StatusCode == http.StatusOK)
	return err
}

// PrepareMetricsBatch function assembles metrics data from agent's storage
// into batch HTTP requests of specified size to send to the server.
func (s *HTTPSender) PrepareMetricsBatch(metrics []dto.Metrics, requests chan *http.Request, batchSize int) {
	batch := make([]dto.Metrics, 0)
	url := s.addressServer + "/updates"
	for i, metric := range metrics {
		batch = append(batch, metric)
		if (i+1)%batchSize == 0 {
			r, err := MetricsPostJson(s.addressClient, s.config.HashKey, s.cipherService, batch, url)
			if err != nil {
				s.logger.Error("batch post response failed", slog.String("error", err.Error()))
				continue
			}
			s.logger.Debug("post metrics batch", slog.Any("response", r))
			requests <- r
			batch = make([]dto.Metrics, 0)
		}
	}
	if len(batch) > 0 {
		if r, err := MetricsPostJson(s.addressClient, s.config.HashKey, s.cipherService, batch, url); err == nil {
			requests <- r
		}
	}
}

// SendMetrics function starts sending of the prepared HTTP requests to metric server.
func (s *HTTPSender) SendMetrics(requests chan *http.Request) {
	s.workerPool.Run(requests, func(req *http.Request) {
		err := util.TryRun(func() (err error) {
			res, err := s.client.Do(req)
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
			s.logger.Error(err.Error())
		}
	})
}

// StartSending function initiates the process of sending collected data from in-memory storage to the metric server.
// If JSON format is supported by the API, the agent will send metrics in batches thus saving on the amount of requests.
func (s *HTTPSender) StartSending(interval time.Duration, retrieve func(names []string) []dto.Metrics) chan<- struct{} {
	if s.updateTick != nil {
		return nil
	}
	quit := make(chan struct{})
	var finishWg sync.WaitGroup
	s.updateTick = time.NewTicker(interval)
	go func() {

	loop:
		for {
			finishWg.Add(1)

			reqs := make(chan *http.Request, s.config.RateLimit)
			go func() {
				mainList := retrieve(s.primaryMetrics)
				secondaryList := retrieve(s.secondaryMetrics)

				if s.JsonAvailable {
					s.PrepareMetricsBatch(mainList, reqs, 8)
				} else {
					s.PrepareMetrics(mainList, reqs)
				}
				s.PrepareMetrics(secondaryList, reqs)
				close(reqs)
				finishWg.Done()
			}()
			go s.SendMetrics(reqs)

			select {
			case t := <-s.updateTick.C:
				s.logger.Debug(fmt.Sprintf("Tick at %s", t.String()))
				continue
			case <-quit:
				s.updateTick.Stop()
				finishWg.Wait()
				break loop

			}
		}
	}()
	return quit
}
