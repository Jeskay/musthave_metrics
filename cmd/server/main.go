package main

import (
	"fmt"
	"musthave_metrics/internal/metric"
	"musthave_metrics/internal/metric/transport"
	"net/http"
	"os"
)

func main() {
	service := metric.NewMetricService()
	handler := transport.NewHandler(service)

	if err := http.ListenAndServe(`:8080`, handler); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
