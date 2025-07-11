package transport

import (
	"html/template"
	"net"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/transport/grpc"
	"github.com/Jeskay/musthave_metrics/internal/metric/transport/http"
)

func RunHTTP(conf *config.ServerConfig, t *template.Template, service *metric.MetricService, onClose func(err error)) {
	r := http.Init(conf, service, t)
	if err := r.Run(conf.Address); err != nil {
		onClose(err)
	}
}

func RunGRPC(conf *config.ServerConfig, service *metric.MetricService, onClose func(err error)) {
	s := grpc.Init(service)
	l, err := net.Listen("tcp", conf.Address)
	if err != nil {
		onClose(err)
	}
	if err := s.Serve(l); err != nil {
		onClose(err)
	}
}
