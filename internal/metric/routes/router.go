package routes

import (
	"html/template"
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/handlers"
	"github.com/gin-gonic/gin"
)

func Init(svc *metric.MetricService, template *template.Template) *gin.Engine {
	r := gin.Default()
	r.SetHTMLTemplate(template)

	v1 := r.Group("/update")
	{
		v1.POST("/counter/:name/:value", handlers.UpdateCounterMetric(svc))
		v1.POST("/gauge/:name/:value", handlers.UpdateGaugeMetric(svc))
		v1.POST("/:type/:name/:value", func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusBadRequest)
		})
	}
	v2 := r.Group("/value")
	{
		v2.GET("/counter/:name", handlers.GetCounterMetric(svc))
		v2.GET("/gauge/:name", handlers.GetGaugeMetric(svc))
		v2.GET("/:type/:name", func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusNotFound)
		})
	}
	r.GET("", handlers.ListMetrics(svc))
	return r
}
