package handlers

import (
	"secret-svc/api/dtos"
	"secret-svc/api/services"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GET - Get Health Handler
// ////////////////////////////
func GetHealthHandler(c *gin.Context) {
	c.JSON(200, dtos.ApiResponse{
		Success: true,
		Message: "Secret Service is Up and Running!",
	})
}

// GET - Get secret by ID Handler
// //////////////////////////////////
func GetSecretHandler(c *gin.Context) {
	id := c.Param("id")
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	version := c.Query("version")

	data, err := services.GetSecret(headers, id, version)

	if err != nil {
		if err == constants.ErrUUIDsNotFound || err == constants.ErrKeyNotFound {
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

	c.JSON(200, dtos.ApiResponse{
		Success: true,
		Message: "Secret Returned",
		Data:    data,
	})
}

// GET - Get secret versions Handler
// ////////////////////////////////////
func GetSecretVersionsHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	id := c.Param("id")

	data, err := services.GetSecretVersions(headers, id)

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
		Message: "Secret Versions Returned",
		Data:    data,
	})
}

// POST - Create Secret Handler
// ///////////////////////////////
func CreateSecretHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	rawRequestBody, _ := utils.ExtractRequestBody(c)
	requestBody, err := dtos.CreateNewSecretReq(rawRequestBody)

	// Invalid Request Body
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	decodedSecret, err := utils.Base64Decode(requestBody.Secret)

	// Invalid base64 error
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error() + " or " + constants.ErrSecretNotBase64Encoded.Error(),
		})
		return
	}

	data, err := services.CreateSecret(headers, decodedSecret)

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
		Message: "New Secret Added",
		Data:    data,
	})
}

// PUT - Update Secret Handler
// ////////////////////////////////
func PutSecretHandler(c *gin.Context) {
	id := c.Param("id")
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	rawRequestBody, _ := utils.ExtractRequestBody(c)
	requestBody, err := dtos.CreateNewSecretReq(rawRequestBody)

	// Invalid request body
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error(),
		})
		return
	}

	decodedSecret, err := utils.Base64Decode(requestBody.Secret)

	// Invalid base64 string
	if err != nil {
		c.JSON(401, dtos.ApiResponse{
			Success: false,
			Message: "ERROR",
			Error:   err.Error() + " or " + constants.ErrSecretNotBase64Encoded.Error(),
		})
		return
	}

	data, err := services.UpdateSecret(headers, id, decodedSecret)

	// Error Updating secret
	if err != nil {
		if err == constants.ErrUUIDsNotFound || err == constants.ErrKeyNotFound {
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
		Message: "Secret Updated",
		Data:    data,
	})
}

// DELETE - Delete Secret Handler
// /////////////////////////////////
func DeleteSecretHandler(c *gin.Context) {
	id := c.Param("id")
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	data, err := services.DeleteSecret(headers, id)

	if err != nil {
		if err == constants.ErrUUIDsNotFound || err == constants.ErrKeyNotFound {
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

	c.JSON(200, dtos.ApiResponse{
		Success: true,
		Message: "Secret Deleted",
		Data:    data,
	})
}

// DELETE - Delete Secret Group Handler
// //////////////////////////////////////////
func DeleteSecretGroupHandler(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	data, err := services.DeleteSecretGroup(headers, headers.ARN, headers.Region)

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
		Message: "Secret Group deleted with all secrets",
		Data:    data,
	})
}
