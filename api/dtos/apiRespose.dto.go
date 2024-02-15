package dtos

//This format is used as the API responce for all API endpoints provided by this service
type ApiResponse struct {
	Success bool        `json:"success" binding:"required"`
	Message string      `json:"message" binding:"required"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}
