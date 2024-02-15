package tests

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"secret-svc/api/dtos"
	"secret-svc/api/handlers"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var secretId string

func init() {
	godotenv.Load("../.env")
	zerolog.SetGlobalLevel(zerolog.Disabled)
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)

	pathParams := []gin.Param{}
	queryParams := url.Values{}

	//Registering to the Secret-svc
	MockJsonPost(ctx, dtos.SystemSecretReq{
		Flow: "SHARED",
	}, pathParams, queryParams, MockSecretHeaders(ctx))
	handlers.CreateSystemSecretHandler(ctx)
}

func TestGetHealthHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	MockJsonGet(ctx, pathParams, queryParams, ctx.Request.Header)

	handlers.GetHealthHandler(ctx)
	assert.EqualValues(t, 200, w.Code)
}

func TestPostSecretHandler(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	ctx.Request.Header.Set(constants.FLOW_HEADER, constants.SHARED_FLOW)

	MockJsonPost(ctx, dtos.SecretReq{
		Secret: "ewogICAgImtleTEiOiAidmFsdWUxIiwKICAgICJrZXkyIjogInZhbHVlMiIKfQ==",
	}, pathParams, queryParams, MockSecretHeaders(ctx))

	handlers.CreateSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 201, w.Code)

	var res dtos.ApiResponse
	utils.DeStringifyJson(w.Body.String(), &res)

	secretId = res.Data.(string)
}

func TestPutSecretHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	ctx.Request.Header.Set(constants.FLOW_HEADER, constants.SHARED_FLOW)
	pathParams := []gin.Param{
		{
			Key:   "id",
			Value: secretId,
		},
	}
	queryParams := url.Values{}

	MockJsonPut(ctx, dtos.SecretReq{
		Secret: "ewogICAgImtleTEiOiAidmFsdWUxIiwKICAgICJrZXkyIjogInZhbHVlMiIKfQ==",
	}, pathParams, queryParams, MockSecretHeaders(ctx))

	handlers.PutSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 201, w.Code)
}

func TestDeleteSecretHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	ctx.Request.Header.Set(constants.FLOW_HEADER, constants.SHARED_FLOW)
	pathParams := []gin.Param{
		{
			Key:   "id",
			Value: secretId,
		},
	}
	queryParams := url.Values{}

	MockJsonDelete(ctx, pathParams, queryParams, MockSecretHeaders(ctx))

	handlers.DeleteSecretHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 200, w.Code)
}

func TestDeleteSecretGroupHandler(t *testing.T) {
	godotenv.Load("../.env")
	w := httptest.NewRecorder()
	ctx := GetTestGinContext(w)
	ctx.Request.Header.Set(constants.FLOW_HEADER, constants.SHARED_FLOW)
	pathParams := []gin.Param{}
	queryParams := url.Values{}

	MockJsonDelete(ctx, pathParams, queryParams, MockSecretHeaders(ctx))

	handlers.DeleteSecretGroupHandler(ctx)
	fmt.Println(w.Body.String())
	assert.EqualValues(t, 200, w.Code)

	//Removing the Secret-svc entry
	MockJsonDelete(ctx, pathParams, queryParams, MockSecretHeaders(ctx))
	handlers.DeleteSystemSecretHandler(ctx)
	fmt.Println(w.Body.String())
}
