package domain

import "errors"

// Setting represents a key-value configuration setting
type Setting struct {
	ID    string      `json:"id" bson:"_id"`
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
	CTime int64       `json:"ctime" bson:"ctime"`
	MTime int64       `json:"mtime" bson:"mtime"`
}

// CreateSettingParams for creating a new setting
type CreateSettingParams struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// UpdateSettingParams for updating a setting
type UpdateSettingParams struct {
	Value interface{} `json:"value" binding:"required"`
}

// Errors
var (
	ErrSettingNotFound      = errors.New("setting not found")
	ErrSettingKeyExists     = errors.New("setting key already exists")
	ErrInvalidSettingKey    = errors.New("invalid setting key")
	ErrInvalidSettingValue  = errors.New("invalid setting value")
)
