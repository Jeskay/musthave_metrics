package routes

import (
	"html/template"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/handlers"
	"github.com/Jeskay/musthave_metrics/internal/metric/middleware"
)

func Init(config *config.ServerConfig, svc *metric.MetricService, template *template.Template) *gin.Engine {
	r := gin.Default()
	r.SetHTMLTemplate(template)
	if config.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(config.TrustedSubnet)
		if err == nil {
			r.Use(middleware.SubnetChecker(subnet))
		}
	}
	r.Use(middleware.Logger(svc.Logger))
	r.Use(middleware.HashDecoder(config.HashKey))
	r.Use(middleware.HashEncoder(config.HashKey))
	r.Use(middleware.GzipDecoder())
	r.Use(middleware.NewGzipHandler().Handle)
	if privateKey, err := config.LoadPrivateKey(); err == nil && privateKey != nil {
		r.Use(middleware.Decipher(privateKey))
	}

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
