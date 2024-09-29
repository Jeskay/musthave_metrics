package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

type CounterMetricHandler struct {
	svc *metric.MetricService
}

func NewCounterMetricHandler(svc *metric.MetricService) CounterMetricHandler {
	return CounterMetricHandler{
		svc: svc,
	}
}

func (h CounterMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		p := strings.Split(r.URL.Path, "/")
		if len(p) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		key := p[0]
		value, err := strconv.ParseInt(p[1], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.svc.SetCounterMetric(key, value)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
