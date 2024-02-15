package main

import (
	"secret-svc/api"
	"secret-svc/api/middlewares"
	"secret-svc/pkg/loggers"
	"secret-svc/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// main function
// //////////////////
func main() {
	// Initializing The Global Logger
	zap.ReplaceGlobals(loggers.GetCustomLogger())
	zap.L().Info("Initializing Secret Service")

	// loading environment variables.
	err := godotenv.Load()
	if err != nil {
		zap.L().Fatal("Error Loading .env file :: " + err.Error())
	}
	zap.L().Info("Environment variables loaded")

	PORT := utils.GetEnvVar("PORT")
	BASE := utils.GetEnvVar("BASE")
	GIN_MODE := utils.GetEnvVar("GIN_MODE")
	REDIS_ADDR := utils.GetEnvVar("REDIS_HOST") + ":" + utils.GetEnvVar("REDIS_PORT")
	REDIS_PASSWORD := utils.GetEnvVar("REDIS_PASSWORD")
	BYPASS_REDIS := utils.GetEnvVar("BYPASS_REDIS")

	// Setting the GIN mode
	if GIN_MODE == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Setting up routes and Middlewares
	router := gin.Default()
	router.Use(middlewares.LogRequests)

	if BYPASS_REDIS != "true" {
		err = middlewares.InitiallizeRedis(REDIS_ADDR, REDIS_PASSWORD)
		if err != nil {
			zap.L().Fatal("Error Initializing Redis :: " + err.Error())
		}

		router.Use(middlewares.RedisLockMiddleware)
	} else {
		zap.L().Info("Redis was bypassed based on the env config")
	}

	api.SetHealthRoute(router)
	router.Use(middlewares.CheckHeaders)
	api.SetSystemSecretRoutes(router)
	router.Use(middlewares.AddDefaultScope)
	api.SetSecretRoutes(router)

	// Listening to Ports
	router.Run(BASE + ":" + PORT)
}
