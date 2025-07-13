package transport

import (
	"html/template"
	"net"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	metric_grpc "github.com/Jeskay/musthave_metrics/internal/metric/transport/grpc"
	metric_http "github.com/Jeskay/musthave_metrics/internal/metric/transport/http"
	"google.golang.org/grpc"
)

func RunHTTP(conf *config.ServerConfig, t *template.Template, service *metric.MetricService, onClose func(err error)) {
	r := metric_http.Init(conf, service, t)
	if err := r.Run(conf.Address); err != nil {
		onClose(err)
	}
}

func RunGRPC(conf *config.ServerConfig, service *metric.MetricService, onClose func(err error), options ...grpc.ServerOption) {
	s := metric_grpc.Init(service, options...)
	l, err := net.Listen("tcp", conf.Address)
	if err != nil {
		onClose(err)
	}
	if err := s.Serve(l); err != nil {
		onClose(err)
	}
}
