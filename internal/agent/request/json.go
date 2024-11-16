package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

func MetricPostJson(hashKey string, metric dto.Metrics, url string) (req *http.Request, err error) {
	var buf bytes.Buffer
	data, err := json.Marshal(metric)
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
	WriteHash(req, buf.Bytes(), hashKey)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}

func MetricsPostJson(hashKey string, metrics []dto.Metrics, url string) (req *http.Request, err error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	data, err := json.Marshal(metrics)
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
	WriteHash(req, buf.Bytes(), hashKey)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}
