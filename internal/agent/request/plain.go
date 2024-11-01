package request

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
)

func MetricPostPlain(name string, metricValue internal.MetricValue, url string) (*http.Request, error) {
	if metricValue.Type == internal.CounterMetric {
		v, ok := metricValue.Value.(int64)
		if !ok {
			v = 0
		}
		url += path.Join(string(metricValue.Type), name, strconv.FormatInt(v, 10))
	} else if metricValue.Type == internal.GaugeMetric {
		v, ok := metricValue.Value.(float64)
		if !ok {
			v = 0
		}
		url += path.Join(string(metricValue.Type), name, strconv.FormatFloat(v, 'f', -1, 64))
	}
	return http.NewRequest(http.MethodPost, url, nil)
}
