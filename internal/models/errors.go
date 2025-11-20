package models

import "errors"

var (
	ErrMissingRequestID = errors.New("missing request_id")
	ErrMissingTenantID  = errors.New("missing tenant_id")
	ErrMissingRoute     = errors.New("missing route")
	ErrMissingModel     = errors.New("missing model")
	ErrMissingTimestamp = errors.New("missing timestamp")
)
