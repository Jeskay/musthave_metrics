package grpc

import (
	"context"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/transport/grpc/interceptors"
	pb "github.com/Jeskay/musthave_metrics/protos"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	metricService *metric.MetricService
	pb.UnimplementedServerServer
}

func (s *Server) GetCounterMetric(ctx context.Context, r *pb.GetCounterRequest) (*pb.GetMetricResponse, error) {
	ok, val := s.metricService.GetCounterMetric(r.Name)
	if !ok {
		return &pb.GetMetricResponse{Error: "not found"}, nil
	}
	return &pb.GetMetricResponse{Metric: &pb.Metric{Id: r.Name, Delta: &val}}, nil
}

func (s *Server) GetGaugeMetric(ctx context.Context, r *pb.GetGaugeRequest) (*pb.GetMetricResponse, error) {
	ok, val := s.metricService.GetGaugeMetric(r.Name)
	if !ok {
		return &pb.GetMetricResponse{Error: "not found"}, nil
	}
	return &pb.GetMetricResponse{Metric: &pb.Metric{Id: r.Name, Value: &val}}, nil
}

func (s *Server) GetMetric(ctx context.Context, r *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	switch r.Metric.Type {
	case string(internal.CounterMetric):
		ok, val := s.metricService.GetCounterMetric(r.Metric.Id)
		if !ok {
			return &pb.GetMetricResponse{Error: "not found"}, nil
		}
		return &pb.GetMetricResponse{Metric: &pb.Metric{Id: r.Metric.Id, Delta: &val}}, nil
	case string(internal.GaugeMetric):
		ok, val := s.metricService.GetGaugeMetric(r.Metric.Id)
		if !ok {
			return &pb.GetMetricResponse{Error: "not found"}, nil
		}
		return &pb.GetMetricResponse{Metric: &pb.Metric{Id: r.Metric.Id, Value: &val}}, nil
	default:
		return &pb.GetMetricResponse{Error: "bad request"}, nil
	}
}

func (s *Server) ListMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.MetricsResponse, error) {
	list, err := s.metricService.GetAllMetrics()
	if err != nil {
		return nil, err
	}
	var metrics []*pb.Metric = make([]*pb.Metric, len(list))
	for i, v := range list {
		metrics[i] = &pb.Metric{Id: v.ID, Delta: v.Delta, Value: v.Value, Type: v.MType}
	}
	return &pb.MetricsResponse{Metrics: metrics}, nil
}

func (s *Server) Ping(ctx context.Context, r *emptypb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{Available: s.metricService.DBHealth()}, nil
}

func (s *Server) UpdateCounterMetric(ctx context.Context, r *pb.UpdateCounterRequest) (*pb.UpdateCounterResponse, error) {
	err := s.metricService.SetCounterMetric(r.Name, r.Value)
	return &pb.UpdateCounterResponse{Error: err.Error()}, nil
}

func (s *Server) UpdateGaugeMetric(ctx context.Context, r *pb.UpdateGaugeRequest) (*pb.UpdateGaugeResponse, error) {
	err := s.metricService.SetGaugeMetric(r.Name, r.Value)
	return &pb.UpdateGaugeResponse{Error: err.Error()}, nil
}

func (s *Server) UpdateMetric(ctx context.Context, r *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	if r.Metric.Type == string(internal.CounterMetric) {
		if err := s.metricService.SetCounterMetric(r.Metric.Id, *r.Metric.Delta); err != nil {
			return &pb.UpdateMetricResponse{Error: err.Error()}, nil
		}
		if ok, v := s.metricService.GetCounterMetric(r.Metric.Id); ok {
			return &pb.UpdateMetricResponse{Value: &pb.Metric{Id: r.Metric.Id, Type: r.Metric.Type, Delta: &v}}, nil
		}
	} else {
		if err := s.metricService.SetGaugeMetric(r.Metric.Id, *r.Metric.Value); err != nil {
			return &pb.UpdateMetricResponse{Error: err.Error()}, nil
		}
		if ok, v := s.metricService.GetGaugeMetric(r.Metric.Id); ok {
			return &pb.UpdateMetricResponse{Value: &pb.Metric{Id: r.Metric.Id, Type: r.Metric.Type, Value: &v}}, nil
		}
	}
	return &pb.UpdateMetricResponse{Error: "failed to update metric"}, nil
}

func (s *Server) UpdateMetrics(ctx context.Context, r *pb.UpdateMetricsRequest) (*pb.MetricsResponse, error) {
	var m = make([]dto.Metrics, len(r.Metrics))
	for i, v := range r.Metrics {
		m[i] = dto.Metrics{ID: v.Id, MType: v.Type, Delta: v.Delta, Value: v.Value}
	}
	metrics := dto.OptimizeMetrics(m)
	if err := s.metricService.SetMetrics(metrics); err != nil {
		return &pb.MetricsResponse{Error: "failed to update metrics"}, nil
	}
	keys := make([]string, len(metrics))
	for i, v := range metrics {
		keys[i] = v.ID
	}
	updated, err := s.metricService.GetMetrics(keys)
	if err != nil {
		return &pb.MetricsResponse{Error: "failed to update metrics"}, nil
	}
	var pbMetrics = make([]*pb.Metric, len(r.Metrics))
	for i, v := range updated {
		pbMetrics[i] = &pb.Metric{Id: v.ID, Type: v.MType, Delta: v.Delta, Value: v.Value}
	}
	return &pb.MetricsResponse{Metrics: pbMetrics}, nil
}

func Init(metricService *metric.MetricService) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.NewLoggingUnaryInterceptor(metricService.Logger)),
	)
	pb.RegisterServerServer(server, &Server{metricService: metricService})
	return server
}
