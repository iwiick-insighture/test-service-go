package dtos

import "net/http"

type RequestLog struct {
	TraceId   string `json:"traceId" binding:"required"`
	Url       string `json:"url" binding:"required"`
	OrgId     string `json:"orgId,omitempty"`
	ProjectId string `json:"projectId,omitempty"`
	Scope     string `json:"scope,omitempty"`
}

// Method for creating new API request Logs
//////////////////////////////////////////////
func CreateNewApiRequestLog(request *http.Request) *RequestLog {
	headers := ExtractCustomHeaders(request.Header)

	return &RequestLog{
		TraceId:   headers.TraceId,
		Url:       request.URL.String(),
		OrgId:     headers.OrgId,
		ProjectId: headers.ProjectId,
		Scope:     headers.Scope,
	}
}
