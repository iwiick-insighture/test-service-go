package middlewares

import (
	"secret-svc/api/dtos"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Middleware to strictly validate Headers
// ////////////////////////////////////////////////
/*
	- Trace Id is required.
	- OrganizationId is required if the ProjectId is provided.
	- ProjectId is required if the Scope is provided.
*/
func CheckHeaders(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)

	//Trace Id is required for the endpoints
	if headers.TraceId == "" {
		zap.L().Error(constants.ErrEmptyTraceId.Error())
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrEmptyTraceId.Error(),
		})
		c.Abort()
	}

	//OrgId is required for the endpoints
	if headers.OrgId == "" {
		zap.L().Error(constants.ErrEmptyOrgId.Error())
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrEmptyOrgId.Error(),
		})
		c.Abort()
	}

	//ProjectId required for the endpoints before scope
	if headers.ProjectId == "" && headers.Scope != "" {
		zap.L().Error(constants.ErrEmptyProjId.Error())
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrEmptyProjId.Error(),
		})
		c.Abort()
	} else if headers.ProjectId != "" && headers.Scope != "" && !utils.ArrayContains(constants.ACCEPTED_SCOPES[:], headers.Scope) {
		// Invalid Scope provided
		zap.L().Error(constants.ErrInvalidScope.Error())
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrInvalidScope.Error(),
		})
		c.Abort()
	}
	c.Next()
}
