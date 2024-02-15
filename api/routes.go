package api

import (
	"secret-svc/api/handlers"
	"secret-svc/api/middlewares"

	"github.com/gin-gonic/gin"
)

var API string = ""

// Health Route
// //////////////////
func SetHealthRoute(router *gin.Engine) {
	healthRouter := router.Group(API)
	healthRouter.GET("/health", handlers.GetHealthHandler)
}

// System Secret Routes
// /////////////////////////
func SetSystemSecretRoutes(router *gin.Engine) {
	systemSecretRouter := router.Group(API + "/system")
	systemSecretRouter.POST("/", handlers.CreateSystemSecretHandler)
	systemSecretRouter.PUT("/", handlers.UpdateSystemSecretHandler)
	systemSecretRouter.DELETE("/", handlers.DeleteSystemSecretHandler)
}

// Secret Routes
// //////////////////
func SetSecretRoutes(router *gin.Engine) {
	secretRouter := router.Group(API + "/secret")
	secretRouter.Use(middlewares.ManageSecretRoutes)
	secretRouter.GET("/:id", handlers.GetSecretHandler)
	secretRouter.GET("/versions/:id", handlers.GetSecretVersionsHandler)
	secretRouter.POST("/", handlers.CreateSecretHandler)
	secretRouter.PUT("/:id", handlers.PutSecretHandler)
	secretRouter.DELETE("/:id", handlers.DeleteSecretHandler)
	secretRouter.DELETE("/group", handlers.DeleteSecretGroupHandler)
}
