package dtos

import (
	"net/http"
	"secret-svc/pkg/constants"
)

// Custom headers needed for the API calls
type CustomHeaders struct {
	OrgId     string `json:"orgId,omitempty"`
	ProjectId string `json:"projectId,omitempty"`
	Scope     string `json:"environmentId,omitempty"`
	Flow      string `json:"flow,omitempty"`
	TraceId   string `json:"traceId" binding:"required"`
	ARN       string `json:"arn,omitempty"`
	Region    string `json:"region,omitempty"`
	Provider  string `json:"provider,omitempty"`
}

// Method for extracting CustomHeaders given the request headers
// ////////////////////////////////////////////////////////////////
func ExtractCustomHeaders(headers http.Header) CustomHeaders {
	return CustomHeaders{
		OrgId:     headers.Get(constants.ORG_ID_HEADER),
		ProjectId: headers.Get(constants.PROJECT_ID_HEADER),
		Scope:     headers.Get(constants.SCOPE_HEADER),
		Flow:      headers.Get(constants.FLOW_HEADER),
		TraceId:   headers.Get(constants.TRACE_ID_HEADER),
		ARN:       headers.Get(constants.ARN_HEADER),
		Region:    headers.Get(constants.REGION_HEADER),
		Provider:  headers.Get(constants.PROVIDER_HEADER),
	}
}
