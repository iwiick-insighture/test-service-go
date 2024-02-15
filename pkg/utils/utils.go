package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"secret-svc/api/dtos"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Helper function to get Env Variables
// //////////////////////////////////////////
func GetEnvVar(key string) string {
	return os.Getenv(key)
}

func GetPrefixedUuid() string {
	uuid, _ := uuid.NewRandom()
	return "secret_" + uuid.String()
}

// Helper function to SetDefaultValus
// /////////////////////////////////////
func SetDefaultIfEmptyValue(value string, defaultValue string) string {
	if value != "" {
		return value
	}

	return defaultValue
}

// Helper method to Extract Request Bodies
// ////////////////////////////////////////////
func ExtractRequestBody(c *gin.Context) (map[string]interface{}, error) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		return nil, err
	}

	return requestBody, nil
}

// Helper function for returning the string prefix
// ///////////////////////////////////////////////////
func CreatePrefix(headers dtos.CustomHeaders) string {
	prefix := ""
	filler := "_"

	if headers.OrgId != "" {
		prefix += headers.OrgId
	}

	if headers.ProjectId != "" {
		prefix += filler
		prefix += headers.ProjectId
	}

	if headers.Scope != "" {
		prefix += filler
		prefix += headers.Scope
	}

	return prefix
}

// Helper function to check if an array containes a specific value
// ///////////////////////////////////////////////////////////////////
func ArrayContains(arr []string, value string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == value {
			return true
		}
	}

	return false
}

// Helper function to stringify Jsons
// //////////////////////////////////////
func StringifyJson(obj interface{}) (string, error) {
	jsonObj, err := json.Marshal(obj)

	if err == nil {
		return string(jsonObj), nil
	}
	return fmt.Sprintf("%+v", string(jsonObj)), nil
}

// Helper function to de-stringify Jsons
// ////////////////////////////////////////
func DeStringifyJson(stringObj string, template interface{}) interface{} {
	return json.Unmarshal([]byte(stringObj), &template)
}

// Helper function for encoding strings to base64
// /////////////////////////////////////////////////
func Base64Encode(value interface{}) string {
	switch v := value.(type) {
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	case string:
		return base64.StdEncoding.EncodeToString([]byte(v))
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return base64.StdEncoding.EncodeToString(jsonBytes)
	}
}

// Helper function for decoding base64 encoded strings
// //////////////////////////////////////////////////////
func Base64Decode(encodedValue string) (string, error) {
	trimmedStr := strings.Trim(encodedValue, `"`)
	decodedStr, err := base64.StdEncoding.DecodeString(trimmedStr)
	if err == nil {
		return string(decodedStr), nil
	}

	return "", err
}
