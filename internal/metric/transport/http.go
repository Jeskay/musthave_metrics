package transport

import (
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/handlers"
)

func NewHandler(svc *metric.MetricService) http.Handler {
	m := http.NewServeMux()
	m.Handle(`/update/gauge/`, http.StripPrefix(`/update/gauge/`, handlers.NewGaugeMetricHandler(svc)))
	m.Handle(`/update/counter/`, http.StripPrefix(`/update/counter/`, handlers.NewCounterMetricHandler(svc)))
	m.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return m
}
