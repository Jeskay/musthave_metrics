package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

func MetricPostJson(hashKey string, cipherService *Cipher, metric dto.Metrics, url string) (req *http.Request, err error) {
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
	if req, err = http.NewRequest(http.MethodPost, url, &buf); err != nil {
		return
	}
	if err := WriteHash(req, buf.Bytes(), hashKey); err != nil {
		return req, err
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return
}

func MetricsPostJson(hashKey string, cipherService *Cipher, metrics []dto.Metrics, url string) (req *http.Request, err error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	ciphered := data
	if cipherService != nil {
		ciphered, err = cipherService.CipherJson(data)
		if err != nil {
			return nil, err
		}
	}
	if _, err := g.Write(ciphered); err != nil {
		return nil, err
	}
	if err = g.Close(); err != nil {
		return nil, err
	}
	if req, err = http.NewRequest(http.MethodPost, url, &buf); err != nil {
		return
	}
	if err := WriteHash(req, buf.Bytes(), hashKey); err != nil {
		return req, err
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if cipherService != nil {
		req.Header.Set("Ciphered", "true")
	}
	return
}
