package middlewares

import (
	"secret-svc/api/dtos"
	"secret-svc/pkg/loggers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var req *dtos.RequestLog

// TODO :: Implement Auditing (Ex: redshift)
// Middleware to Log Requests
// /////////////////////////////////
func LogRequests(c *gin.Context) {
	req = dtos.CreateNewApiRequestLog(c.Request)
	loggers.UpdateLoggerData(req)

	zap.L().Info("Request",
		zap.String("url", req.Url),
	)
	c.Next()
}
