package middlewares

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"secret-svc/api/dtos"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

var redisPool *redis.Pool

// Middleware used for the redis Implementation
// //////////////////////////////////////////////////
func RedisLockMiddleware(c *gin.Context) {
	headers := dtos.ExtractCustomHeaders(c.Request.Header)
	// Adding default scope to lighten traffic
	if headers.ProjectId != "" && headers.Scope == "" {
		c.Request.Header.Set(constants.SCOPE_HEADER, constants.OTHERS_SCOPE)
	}

	lockId := utils.CreatePrefix(headers)

	// Check if the HTTP method is not GET
	if c.Request.Method != http.MethodGet {
		acquiredLock, err := AcquireLock(lockId)

		// Acquire lock
		if acquiredLock {
			zap.L().Info("Lock Accquired for :: " + lockId)
			defer ReleaseLock(lockId)
			c.Next()
		} else {
			zap.L().Info("Failed to acquire Redis lock :: " + lockId + " :: " + err.Error())
			c.AbortWithStatusJSON(503, dtos.ApiResponse{
				Success: false,
				Message: "Failed to acquire Redis lock",
				Error:   err.Error(),
			})
		}
	} else {
		c.Next()
	}
}

// Method for Initializing Redis
// ////////////////////////////////////
func InitiallizeRedis(redisAddr string, redisPassword string) error {
	redisPool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   10,
		IdleTimeout: 30 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisAddr, redis.DialPassword(redisPassword), redis.DialUseTLS(true), redis.DialTLSConfig(&tls.Config{}))
			if err != nil {
				return nil, err
			}
			return conn, err
		},
	}
	// Check Redis liveness
	if err := pingRedis(); err != nil {
		return err
	}

	fmt.Println("Redis Connection Established at :: " + redisAddr)
	return nil
}

// Method for pinging Redis
// ////////////////////////////
func pingRedis() error {
	conn := redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		return err
	}

	return nil
}

// Method for Acquring a Lock
// ///////////////////////////////
func AcquireLock(key string) (bool, error) {
	timeout := 10 * time.Second
	conn := redisPool.Get()
	defer conn.Close()

	lockKey := "secretlock:" + key

	// Calculate the deadline for the lock acquisition attempt
	deadline := time.Now().Add(timeout)

	for {
		// Try to acquire the lock
		reply, err := redis.String(conn.Do("SET", lockKey, 1, "EX", int(timeout.Seconds()), "NX"))
		if err == nil && reply == "OK" {
			// Lock acquired successfully
			return true, nil
		}

		// Check if the deadline for the lock acquisition attempt has passed
		if time.Now().After(deadline) {
			return false, errors.New("timeout: unable to acquire lock")
		}

		// Sleep for a short duration before the next attempt
		time.Sleep(100 * time.Millisecond)
	}
}

// Method for releazing Locks
// ///////////////////////////////
func ReleaseLock(key string) {
	conn := redisPool.Get()
	defer conn.Close()

	lockKey := "secretlock:" + key
	conn.Do("DEL", lockKey)
	fmt.Println("Lock Released for :: " + key)
}
