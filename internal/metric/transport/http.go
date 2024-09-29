package transport

import (
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/handlers"
)

func NewHandler(svc *metric.MetricService) http.Handler {
	m := http.NewServeMux()
	m.Handle(`/update/{type}/{name}/{value}`, handlers.NewUpdateMetricHandler(svc))
	m.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return m
}
