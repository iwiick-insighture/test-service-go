package dtos

import (
	"fmt"
	"secret-svc/pkg/constants"
	"strings"
)

type SystemSecretReq struct {
	Flow     string `json:"flow,omitempty"`
	ARN      string `json:"arn,omitempty"`
	Region   string `json:"region,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type SecretReq struct {
	Secret string `json:"secret,omitempty"`
}

// Helper method for creating a System Secret Obj
// ///////////////////////////////////////////////////
func CreateNewSystemSecretReq(body interface{}) (SystemSecretReq, error) {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return SystemSecretReq{}, constants.ErrFormat
	}

	flow, ok := bodyMap[strings.ToLower(constants.FLOW_META_DATA)].(string)
	if !ok {
		return SystemSecretReq{}, constants.ErrMissingFlowAttr
	}

	var arn, region, provider string
	// Check if the flow is "PRIVATE" to extract ARN, Region, and Provider
	if flow == constants.PRIVATE_FLOW {
		var missingAttrs []string

		if arnVal, ok := bodyMap[strings.ToLower(constants.ARN_META_DATA)].(string); ok {
			arn = arnVal
		} else {
			missingAttrs = append(missingAttrs, strings.ToLower(constants.ARN_META_DATA))
		}

		if regionVal, ok := bodyMap[strings.ToLower(constants.REGION_META_DATA)].(string); ok {
			region = regionVal
		} else {
			missingAttrs = append(missingAttrs, strings.ToLower(constants.REGION_META_DATA))
		}

		if providerVal, ok := bodyMap[strings.ToLower(constants.PROVIDER_META_DATA)].(string); ok {
			provider = providerVal
		} else {
			missingAttrs = append(missingAttrs, strings.ToLower(constants.PROVIDER_META_DATA))
		}

		if len(missingAttrs) > 1 {
			missingValues := strings.Join(missingAttrs, ", ")
			return SystemSecretReq{}, fmt.Errorf("missing %s for 'Private' flow", missingValues)
		} else if len(missingAttrs) == 1 {
			return SystemSecretReq{}, fmt.Errorf("missing %s for 'Private' flow", missingAttrs[0])
		}
	}

	return SystemSecretReq{
		Flow:     flow,
		ARN:      arn,
		Region:   region,
		Provider: provider,
	}, nil
}

// Helper method for creating a new secret request
// /////////////////////////////////////////////////////
func CreateNewSecretReq(body interface{}) (SecretReq, error) {
	bodyMap, _ := body.(map[string]interface{})
	secret, ok := bodyMap[strings.ToLower(constants.SECRET_META_DATA)].(string)
	if !ok {
		return SecretReq{}, constants.ErrMissingSecretAttr
	}

	return SecretReq{
		Secret: secret,
	}, nil
}
