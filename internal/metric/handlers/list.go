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

func NewMetricString(metric *internal.Metric) MetricString {
	mStr := MetricString{
		Name: metric.Name,
		Type: string(metric.Value.Type),
	}
	if metric.Value.Type == internal.CounterMetric {
		v := metric.Value.Value.(int64)
		mStr.Value = strconv.FormatInt(v, 10)
	} else {
		v := metric.Value.Value.(float64)
		mStr.Value = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return mStr
}

func ListMetrics(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		list := svc.GetAllMetrics()
		metrics := make([]MetricString, len(list))
		for i, m := range list {
			metrics[i] = NewMetricString(m)
		}

		c.HTML(http.StatusOK, "/templates/list.tmpl", gin.H{
			"Metrics": metrics,
			"title":   "List of Metrics",
		})
	}
}
