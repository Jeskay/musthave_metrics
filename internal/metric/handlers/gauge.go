package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

type GaugeMetricHandler struct {
	svc *metric.MetricService
}

func NewGaugeMetricHandler(svc *metric.MetricService) GaugeMetricHandler {
	return GaugeMetricHandler{
		svc: svc,
	}
}

func (h GaugeMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		p := strings.Split(r.URL.Path, "/")
		if len(p) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		key := p[0]
		value, err := strconv.ParseFloat(p[1], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.svc.SetGaugeMetric(key, value)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
