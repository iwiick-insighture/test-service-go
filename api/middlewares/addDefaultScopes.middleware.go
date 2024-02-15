package middlewares

import (
	"secret-svc/api/dtos"
	"secret-svc/pkg/constants"

	"github.com/gin-gonic/gin"
)

// Middleware to add Default Scopes
// ////////////////////////////////////////////////
func AddDefaultScope(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)

	// Adding the "OTHERS" scope if the projectid is provided without scopes
	if headers.ProjectId != "" && headers.Scope == "" {
		c.Request.Header.Set(constants.SCOPE_HEADER, constants.OTHERS_SCOPE)
	}
	c.Next()
}
