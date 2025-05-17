package routes

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/handlers"
	"github.com/Jeskay/musthave_metrics/internal/metric/middleware"
)

func Init(hashKey string, svc *metric.MetricService, template *template.Template) *gin.Engine {
	r := gin.Default()
	r.SetHTMLTemplate(template)
	r.Use(middleware.Logger(svc.Logger))
	r.Use(middleware.HashDecoder(hashKey))
	r.Use(middleware.HashEncoder(hashKey))
	r.Use(middleware.GzipDecoder())
	r.Use(middleware.NewGzipHandler().Handle)

	v1 := r.Group("/update")
	{
		v1.POST("/", handlers.UpdateMetricJson(svc))
		v1.POST("/counter/:name/:value", handlers.UpdateCounterMetricRaw(svc))
		v1.POST("/gauge/:name/:value", handlers.UpdateGaugeMetricRaw(svc))
		v1.POST("/:type/:name/:value", func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusBadRequest)
		})

	}
	v2 := r.Group("/value")
	{
		v2.POST("/", handlers.GetMetricJson(svc))
		v2.GET("/counter/:name", handlers.GetCounterMetric(svc))
		v2.GET("/gauge/:name", handlers.GetGaugeMetric(svc))
		v2.GET("/:type/:name", func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusNotFound)
		})

	}
	r.POST("/updates", handlers.UpdateMetricsJson(svc))
	r.GET("/ping", handlers.Ping(svc))
	r.GET("", handlers.ListMetrics(svc))
	return r
}
