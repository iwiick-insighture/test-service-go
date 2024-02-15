package middlewares

import (
	"secret-svc/api/dtos"
	"secret-svc/api/services"
	"secret-svc/pkg/constants"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Middleware to Strictly manage API routes based on client registrations
// //////////////////////////////////////////////////////////////////////////
func ManageSecretRoutes(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	version := c.Query("version")

	// Get system secret including ARN, Region, and Provider
	_, arn, region, provider, flow, err := services.GetSystemSecret(headers, version)
	// Check registrations
	if err != nil {
		if err == constants.ErrUnregisteredKey {
			zap.L().Error(err.Error())
			c.JSON(401, dtos.ApiResponse{
				Success: false,
				Message: "UNAUTHORIZED",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// For other errors
		zap.L().Error(err.Error())
		c.JSON(503, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		c.Abort()
		return
	}

	// Set ARN, Region, and Provider to the headers
	c.Request.Header.Set(constants.ARN_HEADER, arn)
	c.Request.Header.Set(constants.REGION_HEADER, region)
	c.Request.Header.Set(constants.PROVIDER_HEADER, provider)

	// Set flow to the header
	c.Request.Header.Set(constants.FLOW_HEADER, flow)

	c.Next()
}
