package handlers

import (
	"secret-svc/api/dtos"
	"secret-svc/api/services"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GET - Get System Secret Handler
// ////////////////////////////////////
func GetSystemSecretHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	data, _, _, _, _, err := services.GetSystemSecret(headers, "")

	if err != nil {
		c.JSON(503, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(200, dtos.ApiResponse{
		Success: true,
		Message: "System Secret Retrived",
		Data:    data,
	})
}

// POST - Create System Secret Handler
// //////////////////////////////////////////
func CreateSystemSecretHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	rawRequestBody, _ := utils.ExtractRequestBody(c)
	requestBody, err := dtos.CreateNewSystemSecretReq(rawRequestBody)

	// Invalid request Body
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	// Invalid flow type
	if !utils.ArrayContains(constants.ACCEPTED_FLOWS[:], requestBody.Flow) {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrInvalidFlow.Error(),
		})
		return
	}

	//Check if Provider, ARN, and Region are missing
	if requestBody.Flow == constants.PRIVATE_FLOW {
		if requestBody.Provider == "" || requestBody.ARN == "" || requestBody.Region == "" {
			c.JSON(401, dtos.ApiResponse{
				Success: false,
				Message: "ERROR",
				Error:   constants.ErrEmptyPvtFlowData.Error(),
			})
			return
		}
	}

	data, err := services.CreateSystemSecret(headers, requestBody)

	// Error Creating System secret
	if err != nil {
		c.JSON(503, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}
	c.JSON(201, dtos.ApiResponse{
		Success: true,
		Message: "New System Secret Added",
		Data:    data,
	})
}

// PUT - Update System Secret Handler
// /////////////////////////////////////////
func UpdateSystemSecretHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	rawRequestBody, _ := utils.ExtractRequestBody(c)
	requestBody, err := dtos.CreateNewSystemSecretReq(rawRequestBody)

	// Invalid Request Body
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	// Invalid Flow type
	if !utils.ArrayContains(constants.ACCEPTED_FLOWS[:], requestBody.Flow) {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   constants.ErrInvalidFlow.Error(),
		})
		return
	}

	newIDs, err := services.UpdateSystemSecret(headers, requestBody)
	// Error updating System Secret
	if err != nil {
		if err == constants.ErrKeyNotFound || err == constants.ErrUnexpectedFlow {
			c.JSON(404, dtos.ApiResponse{
				Success: false,
				Message: "ERROR",
				Error:   err.Error(),
			})
			return
		}

		c.JSON(503, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return

	}

	c.JSON(201, dtos.ApiResponse{
		Success: true,
		Message: "System Secret Updated. Migrated Ids...",
		Data:    newIDs,
	})
}

// DELETE - Delete System Secret Handler
// //////////////////////////////////////////
func DeleteSystemSecretHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	secretNames, err := services.DeleteSystemSecret(headers)

	if err != nil {
		c.JSON(503, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(200, dtos.ApiResponse{
		Success: true,
		Message: "Deleted Secrets from the System Secret Manager",
		Data:    secretNames,
	})
}
