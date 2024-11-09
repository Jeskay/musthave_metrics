package dto

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

func FilterUniqueMetrics(metrics []Metrics) []Metrics {
	m_metrics := make(map[string]Metrics)
	for _, v := range metrics {
		m_metrics[v.ID] = v
	}
	res := make([]Metrics, len(m_metrics))
	i := 0
	for _, v := range m_metrics {
		res[i] = v
		i++
	}
	return res
}
