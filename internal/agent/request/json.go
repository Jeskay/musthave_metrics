package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

func MetricPostJson(name string, metricValue internal.MetricValue, url string) (req *http.Request, err error) {
	var metrics dto.Metrics
	var buf bytes.Buffer
	if metricValue.Type == internal.CounterMetric {
		metrics = dto.NewCounterMetrics(name, metricValue.Value.(int64))
	} else if metricValue.Type == internal.GaugeMetric {
		metrics = dto.NewGaugeMetrics(name, metricValue.Value.(float64))
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(data); err != nil {
		return nil, err
	}
	if err = g.Close(); err != nil {
		return nil, err
	}
	req, err = http.NewRequest(http.MethodPost, url, &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}

func MetricsPostJson(metrics []*internal.Metric, url string) (req *http.Request, err error) {
	var buf bytes.Buffer
	var metricsJson []dto.Metrics = make([]dto.Metrics, len(metrics))
	g := gzip.NewWriter(&buf)

	for i, m := range metrics {
		var metric dto.Metrics
		if m.Value.Type == internal.CounterMetric {
			metric = dto.NewCounterMetrics(m.Name, m.Value.Value.(int64))
		} else if m.Value.Type == internal.GaugeMetric {
			metric = dto.NewGaugeMetrics(m.Name, m.Value.Value.(float64))
		}
		metricsJson[i] = metric
	}
	data, err := json.Marshal(metricsJson)
	if err != nil {
		return nil, err
	}
	if _, err := g.Write(data); err != nil {
		return nil, err
	}
	if err = g.Close(); err != nil {
		return nil, err
	}
	req, err = http.NewRequest(http.MethodPost, url, &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}
