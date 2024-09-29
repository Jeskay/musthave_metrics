package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Jeskay/musthave_metrics/internal/metric/transport"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

func main() {
	service := metric.NewMetricService()
	handler := transport.NewHandler(service)

	if err := http.ListenAndServe(`:8080`, handler); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
