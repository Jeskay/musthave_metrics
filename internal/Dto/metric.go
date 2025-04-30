package dto

import "database/sql"

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewCounterMetrics(name string, value int64) Metrics {
	return Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
}

func NewGaugeMetrics(name string, value float64) Metrics {
	return Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

func (metric Metrics) QueryValues() (counter sql.NullInt64, gauge sql.NullFloat64) {
	if metric.MType == "gauge" {
		counter.Valid = false
		if metric.Value == nil {
			gauge.Valid = false
			return
		}
		gauge.Valid = true
		gauge.Float64 = *metric.Value
	}
	if metric.MType == "counter" {
		gauge.Valid = false
		if metric.Delta == nil {
			counter.Valid = false
			return
		}
		counter.Valid = true
		counter.Int64 = *metric.Delta
	}
	return
}

func OptimizeMetrics(metrics []Metrics) []Metrics {
	mMetrics := make(map[string]Metrics, len(metrics))
	for _, v := range metrics {
		value, exists := mMetrics[v.ID]
		if exists && v.MType == "counter" && value.Delta != nil && v.Delta != nil {
			*v.Delta = *v.Delta + *value.Delta
		}
		mMetrics[v.ID] = v
	}
	res := make([]Metrics, len(mMetrics))
	i := 0
	for _, v := range mMetrics {
		res[i] = v
		i++
	}
	return res
}
