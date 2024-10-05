package main

import (
	"github.com/Jeskay/musthave_metrics/internal/metric/routes"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

func main() {
	service := metric.NewMetricService()
	r := routes.Init(service)

	r.Run(`:8080`)
}
