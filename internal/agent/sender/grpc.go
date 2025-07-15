package sender

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/util"
	"github.com/Jeskay/musthave_metrics/pkg/worker"
	pb "github.com/Jeskay/musthave_metrics/protos"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCSender struct {
	updateTick       *time.Ticker
	workerPool       *worker.WorkerPool[*pb.Metric]
	client           pb.ServerClient
	addressClient    net.IP
	logger           *slog.Logger
	config           *config.AgentConfig
	primaryMetrics   []string
	secondaryMetrics []string
}

func NewGRPCSender(client pb.ServerClient, conf *config.AgentConfig, logger slog.Handler, primaryMetrics []string, secondaryMetrics []string) *GRPCSender {
	sender := &GRPCSender{
		client:           client,
		config:           conf,
		logger:           slog.New(logger),
		workerPool:       worker.NewWorkerPool[*pb.Metric](conf.RateLimit),
		primaryMetrics:   primaryMetrics,
		secondaryMetrics: secondaryMetrics,
	}
	ip, err := util.GetSelfIP()
	if err != nil {
		sender.logger.Error("unknown issue when establishing logical connection to DNS", slog.Attr{Key: "error", Value: slog.AnyValue(err)})
	}
	sender.addressClient = ip
	return sender
}

// PrepareMetrics function converts metrics data to protobuf format to send to the server.
func (s *GRPCSender) PrepareMetrics(metrics []dto.Metrics, protos chan *pb.Metric) {
	var wg sync.WaitGroup
	for _, metric := range metrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			protos <- &pb.Metric{Id: metric.ID, Type: metric.MType, Delta: metric.Delta, Value: metric.Value}
		}()
	}
	wg.Wait()
}

// CheckAPIAvailability of the metric server and returns error if it is unaccessible.
func (s *GRPCSender) CheckAPIAvailability() error {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "uri", s.addressClient.String())
	res, err := s.client.Ping(ctx, &emptypb.Empty{})
	if err == nil && !res.Available {
		return errors.New("server unavailable")
	}
	return err
}

// SendMetrics function starts sending of the prepared metric data to the server via GRPC.
func (s *GRPCSender) SendMetrics(metrics chan *pb.Metric) {
	s.workerPool.Run(metrics, func(m *pb.Metric) {
		err := util.TryRun(func() (err error) {
			ctx := metadata.AppendToOutgoingContext(context.Background(), "uri", s.addressClient.String())
			res, err := s.client.UpdateMetric(ctx, &pb.UpdateMetricRequest{Metric: m})
			if err != nil {
				return
			}
			if res != nil {
				s.logger.Info("Response", slog.String("Value", res.Value.String()))
			}
			return
		}, util.IsConnectionRefused)
		if err != nil {
			s.logger.Error(err.Error())
		}
	})
}

// StartSending function initiates the process of sending collected data from in-memory storage to the metric server.
func (s *GRPCSender) StartSending(interval time.Duration, retrieve func(names []string) []dto.Metrics) chan<- struct{} {
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
			reqs := make(chan *pb.Metric, s.config.RateLimit)
			go func() {
				s.PrepareMetrics(retrieve(s.primaryMetrics), reqs)
				s.PrepareMetrics(retrieve(s.secondaryMetrics), reqs)
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
