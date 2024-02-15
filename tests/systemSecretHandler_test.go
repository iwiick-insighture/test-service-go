package tests

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"secret-svc/api/dtos"
	"secret-svc/api/handlers"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPostSystemSecretHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	MockJsonPost(ctx, dtos.SystemSecretReq{
		Flow: "SHARED",
	}, pathParams, queryParams, MockSystemSecretHeaders(ctx))

	handlers.CreateSystemSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 201, w.Code)
}

func TestPutSystemSecretHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	MockJsonPut(ctx, dtos.SystemSecretReq{
		Flow: "SHARED",
	}, pathParams, queryParams, MockSystemSecretHeaders(ctx))

	handlers.UpdateSystemSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 201, w.Code)
}

func TestDeleteSystemSecretGroupHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	MockJsonDelete(ctx, pathParams, queryParams, MockSystemSecretHeaders(ctx))

	handlers.DeleteSystemSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 200, w.Code)
}
