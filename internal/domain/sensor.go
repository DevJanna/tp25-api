package domain

import (
	"errors"
	"time"
	"tp25-api/lib"
)

type Sensor struct {
	ID      string   `json:"id" bson:"_id"`
	Code    string   `json:"code" bson:"code"`
	Alias   *string  `json:"alias,omitempty" bson:"alias,omitempty"`
	Name    string   `json:"name" bson:"name"`
	Metrics []string `json:"metrics" bson:"metrics"`
	CTime   int64    `json:"ctime" bson:"ctime"`
	MTime   int64    `json:"mtime" bson:"mtime"`
	DTime   *int64   `json:"dtime,omitempty" bson:"dtime,omitempty"`
}

type CreateSensorParams struct {
	Code  string  `json:"code" binding:"required"`
	Alias *string `json:"alias"`
	Name  string  `json:"name" binding:"required"`
}

type UpdateSensorParams struct {
	Code  *string `json:"code"`
	Alias *string `json:"alias"`
	Name  *string `json:"name"`
}

type Range struct {
	Min  float64 `json:"min" bson:"min"`
	Max  float64 `json:"max" bson:"max"`
	Code string  `json:"code" bson:"code"`
}

type Metric struct {
	ID    string  `json:"id" bson:"_id"`
	Code  string  `json:"code" bson:"code"`
	Name  string  `json:"name" bson:"name"`
	Unit  string  `json:"unit" bson:"unit"`
	Alias *string `json:"alias,omitempty" bson:"alias,omitempty"`
	Range []Range `json:"range,omitempty" bson:"range,omitempty"`
	CTime int64   `json:"ctime" bson:"ctime"`
	MTime int64   `json:"mtime" bson:"mtime"`
	DTime *int64  `json:"dtime,omitempty" bson:"dtime,omitempty"`
}

type CreateMetricParams struct {
	Alias *string `json:"alias"`
	Unit  string  `json:"unit" binding:"required"`
	Code  string  `json:"code" binding:"required"`
	Name  string  `json:"name" binding:"required"`
	Range []Range `json:"range"`
}

type UpdateMetricParams struct {
	Unit  *string `json:"unit"`
	Code  *string `json:"code"`
	Name  *string `json:"name"`
	Range []Range `json:"range"`
}

// Record represents a sensor data record with dynamic metric fields
// _id: timestamp in seconds
// c: server create timestamp in milliseconds
// Other fields are dynamic metric values (e.g., WAU, DR, V, Q, Q_of)
type Record map[string]interface{}

// GetTimestamp returns the sensor timestamp (_id field) in seconds
func (r Record) GetTimestamp() int64 {
	if t, ok := r["_id"].(int32); ok {
		return int64(t)
	}
	if t, ok := r["id"].(int32); ok {
		return int64(t)
	}

	return 0
}

// GetCreateTime returns the server create timestamp (c field) in milliseconds
func (r Record) GetCreateTime() int64 {
	if c, ok := r["c"].(int64); ok {
		return c
	}
	return 0
}

// GetFloat gets a float value from the record
func (r Record) GetFloat(key string) float64 {
	if v, ok := r[key].(float64); ok {
		return v
	}
	if v, ok := r[key].(int64); ok {
		return float64(v)
	}
	if v, ok := r[key].(int); ok {
		return float64(v)
	}
	return 0
}

type QueryRecord struct {
	Time  []int64 `json:"time" form:"time"`
	Limit *int    `json:"limit" form:"limit"`
	Skip  *int    `json:"skip" form:"skip"`
}

type RecordsResult struct {
	Records []Record
	Total   int64
}

type ExportType string

const (
	ExportExcel  ExportType = "export_excel"
	ExportCSV    ExportType = "export_csv"
	ExportReport ExportType = "report"
)

type DailyReport struct {
	Date  string             `json:"date" bson:"date"`
	Avg   map[string]float64 `json:"avg" bson:"avg"`
	Min   map[string]float64 `json:"min" bson:"min"`
	Max   map[string]float64 `json:"max" bson:"max"`
	Count int                `json:"count" bson:"count"`
}

var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrMetricCodeExisted  = errors.New("metric code existed")
	ErrMetricMustHaveCode = errors.New("metric must have code")
	ErrRecordIDExisted    = errors.New("record id existed")
)

// NewMetric creates a new metric with timestamps
func NewMetric(params CreateMetricParams) *Metric {
	now := time.Now().UnixMilli()
	return &Metric{
		ID:    lib.Rand.Char(12),
		Code:  params.Code,
		Name:  params.Name,
		Unit:  params.Unit,
		Alias: params.Alias,
		Range: params.Range,
		CTime: now,
		MTime: now,
	}
}
