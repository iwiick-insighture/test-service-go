package constants

import (
	"errors"
	"fmt"
)

var ErrFormat = errors.New("invalid request format. expected a json object with key-value pairs")
var ErrInvalidFlow = fmt.Errorf("invalid 'flow' type in request body. the flow can be '%s' or '%s'", ACCEPTED_FLOWS[0], ACCEPTED_FLOWS[1])
var ErrInvalidScope = fmt.Errorf("invalid 'scope' in headers. x-scope can be  '%s', '%s' or '%s'", ACCEPTED_SCOPES[0], ACCEPTED_SCOPES[1], ACCEPTED_SCOPES[2])
var ErrMissingFlowAttr = errors.New("'flow' attribute missing or not a string in request body")
var ErrMissingSecretAttr = errors.New("'secret' attribute missing or not a string in request body")
var ErrSecretNotBase64Encoded = errors.New("'secret' attribute value might not be base64 encoded")
var ErrEmptyOrgId = errors.New("organization id cannot be empty. check headers")
var ErrEmptyProjId = errors.New("scope can't exists without a project. check headers")
var ErrEmptyTraceId = errors.New("trace id cannot be empty. check headers")
var ErrEmptyPvtFlowData = errors.New("missing values for attributes in request body")
var ErrUnexpectedFlow = errors.New("invalid migration attempt. check values for flow")
var ErrOrgNotFound = errors.New("organization not found")
var ErrKeyNotFound = errors.New("secret key not found check headers")
var ErrUUIDsNotFound = errors.New("uuid not found ")
var ErrSecretsNotFound = errors.New("provided key does not have secrets for migration")
var ErrKeyExsists = errors.New("key already exsist  check headers")
var ErrUnregisteredKey = errors.New("provided key is not registered to use the secret service  check headers")
var ErrInvalidMigration = errors.New("invalid migration attempt. migrations can only be done from shared->external or external->external")
