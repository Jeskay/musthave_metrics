package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

type UpdateMetricHandler struct {
	svc *metric.MetricService
}

func NewUpdateMetricHandler(svc *metric.MetricService) UpdateMetricHandler {
	return UpdateMetricHandler{
		svc: svc,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		metricType := r.PathValue("type")
		metricName := r.PathValue("name")
		v := r.PathValue("value")
		if metricName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if metricType == "counter" {
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			h.svc.SetCounterMetric(metricName, value)
		} else if metricType == "gauge" {
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			h.svc.SetGaugeMetric(metricName, value)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
