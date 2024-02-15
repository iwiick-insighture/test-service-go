package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"secret-svc/pkg/constants"

	"github.com/gin-gonic/gin"
)

func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}

	return ctx
}

// mock GET request
func MockJsonGet(c *gin.Context, pathParams gin.Params, queryParams url.Values, headers map[string][]string) {
	c.Request.Method = "GET"
	c.Request.Header = headers
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = pathParams
	c.Request.URL.RawQuery = queryParams.Encode()
}

func MockJsonPost(c *gin.Context, content interface{}, pathParams gin.Params, queryParams url.Values, headers map[string][]string) {
	c.Request.Method = "POST"
	c.Request.Header = headers
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = pathParams
	c.Request.URL.RawQuery = queryParams.Encode()

	jsonbytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
}

func MockJsonPut(c *gin.Context, content interface{}, pathParams gin.Params, queryParams url.Values, headers map[string][]string) {
	c.Request.Method = "PUT"
	c.Request.Header = headers
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = pathParams
	c.Request.URL.RawQuery = queryParams.Encode()

	jsonbytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
}

func MockJsonDelete(c *gin.Context, pathParams gin.Params, queryParams url.Values, headers map[string][]string) {
	c.Request.Method = "DELETE"
	c.Request.Header = headers
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = pathParams
	c.Request.URL.RawQuery = queryParams.Encode()
}

func MockSystemSecretHeaders(c *gin.Context) http.Header {
	headers := c.Request.Header
	headers.Set(constants.ORG_ID_HEADER, "test-system-111")
	headers.Set(constants.PROJECT_ID_HEADER, "test-system-222")
	headers.Set(constants.SCOPE_HEADER, "test-system-333")
	headers.Set(constants.TRACE_ID_HEADER, "test-system-444")

	return headers
}

func MockSecretHeaders(c *gin.Context) http.Header {
	headers := c.Request.Header
	headers.Set(constants.ORG_ID_HEADER, "test-111")
	headers.Set(constants.PROJECT_ID_HEADER, "test-222")
	headers.Set(constants.SCOPE_HEADER, "test-333")
	headers.Set(constants.TRACE_ID_HEADER, "test-444")
	headers.Set(constants.FLOW_HEADER, "SHARED")

	return headers
}

func ExtractRequestBody(body []byte) (map[string]interface{}, error) {
	var responseBody map[string]interface{}
	err := json.Unmarshal(body, &responseBody)

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
