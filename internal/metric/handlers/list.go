package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

type MetricString struct {
	Name  string
	Type  string
	Value string
}

func ListMetrics(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		list := svc.GetAllMetrics()
		metrics := make([]MetricString, len(list))
		for i, m := range list {
			if m.Value.Type == internal.CounterMetric {
				v := m.Value.Value.(int64)
				metrics[i] = MetricString{
					Name:  m.Name,
					Type:  string(m.Value.Type),
					Value: strconv.FormatInt(v, 10),
				}
			} else if m.Value.Type == internal.GaugeMetric {
				v := m.Value.Value.(float64)
				metrics[i] = MetricString{
					Name:  m.Name,
					Type:  string(m.Value.Type),
					Value: strconv.FormatFloat(v, 'f', -1, 64),
				}
			}
		}

		c.HTML(http.StatusOK, "list.tmpl", gin.H{
			"Metrics": metrics,
			"title":   "List of Metrics",
		})
	}
}
