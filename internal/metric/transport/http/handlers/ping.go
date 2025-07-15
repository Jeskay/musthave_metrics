package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

// Ping handles ping request.
//
//	Method: GET
//	Endpoint: /ping
//
// Example usage with curl:
//
//	curl -X GET http://localhost:9009/ping
//
//	On success, returns HTTP 200 OK.
//	On fail, returns HTTP 500 Internal Server Error.
func Ping(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc.DBHealth() {
			c.Writer.WriteHeader(http.StatusOK)
		} else {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
	}
}
