package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

func NewPlainPost(name string, metricValue internal.MetricValue, url string) (*http.Request, error) {
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

func NewJsonPost(name string, metricValue internal.MetricValue, url string) (req *http.Request, err error) {
	var metrics dto.Metrics
	if metricValue.Type == internal.CounterMetric {
		metrics = dto.NewCounterMetrics(name, metricValue.Value.(int64))
	} else if metricValue.Type == internal.GaugeMetric {
		metrics = dto.NewGaugeMetrics(name, metricValue.Value.(float64))
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}
