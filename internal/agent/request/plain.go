package request

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

func MetricPostPlain(name string, metric dto.Metrics, url string) (*http.Request, error) {
	if internal.MetricType(metric.MType) == internal.CounterMetric && metric.Delta != nil {
		url += path.Join(metric.MType, name, strconv.FormatInt(*metric.Delta, 10))
	} else if internal.MetricType(metric.MType) == internal.GaugeMetric && metric.Value != nil {
		url += path.Join(metric.MType, name, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
	}
	return http.NewRequest(http.MethodPost, url, nil)
}
