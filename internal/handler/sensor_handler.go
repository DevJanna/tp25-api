package handler

import (
	"net/http"
	"strconv"

	"tp25-api/internal/domain"
	"tp25-api/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type SensorHandler struct {
	service *service.SensorService
}

func NewSensorHandler(service *service.SensorService) *SensorHandler {
	return &SensorHandler{service: service}
}

// Metric endpoints

// ListMetrics godoc
// @Summary List all metrics
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /metrics [get]
func (h *SensorHandler) ListMetrics(c *gin.Context) {
	pagination := domain.ParsePaginationParams(c)

	// Build filter
	filter := bson.M{}

	metrics, total, err := h.service.ListMetricsWithPagination(c.Request.Context(), pagination, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := domain.NewPaginatedResponse(metrics, pagination.Page, pagination.PageSize, total, filter)
	c.JSON(http.StatusOK, response)
}

// GetMetric godoc
// @Summary Get metric by ID
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param id path string true "Metric ID"
// @Success 200 {object} domain.Metric
// @Failure 404 {object} map[string]interface{}
// @Router /metrics/{id} [get]
func (h *SensorHandler) GetMetric(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	metric, err := h.service.GetMetric(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		if err == domain.ErrMetricNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metric)
}

// CreateMetric godoc
// @Summary Create a new metric
// @Tags metrics
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateMetricParams true "Metric data"
// @Success 201 {object} domain.Metric
// @Failure 409 {object} map[string]interface{}
// @Router /metrics [post]
func (h *SensorHandler) CreateMetric(c *gin.Context) {
	var params domain.CreateMetricParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric, err := h.service.CreateMetric(c.Request.Context(), params)
	if err != nil {
		if err == domain.ErrMetricCodeExisted {
			c.JSON(http.StatusConflict, gin.H{"error": "metric code already exists"})
			return
		}
		if err == domain.ErrMetricMustHaveCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metric must have code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, metric)
}

// UpdateMetric godoc
// @Summary Update metric
// @Tags metrics
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Metric ID"
// @Param request body domain.UpdateMetricParams true "Update data"
// @Success 200 {object} domain.Metric
// @Failure 404 {object} map[string]interface{}
// @Router /metrics/{id} [put]
func (h *SensorHandler) UpdateMetric(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var params domain.UpdateMetricParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric, err := h.service.UpdateMetric(c.Request.Context(), id, params)
	if err != nil {
		if err == domain.ErrMetricNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metric)
}

// DeleteMetric godoc
// @Summary Delete metric (soft delete)
// @Tags metrics
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Metric ID"
// @Success 200 {object} domain.Metric
// @Failure 404 {object} map[string]interface{}
// @Router /metrics/{id} [delete]
func (h *SensorHandler) DeleteMetric(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	metric, err := h.service.DeleteMetric(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrMetricNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metric)
}

// Record endpoints

// ListRecords godoc
// @Summary List sensor records for a box
// @Tags boxes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Box ID"
// @Param time_min query int false "Min timestamp (seconds)"
// @Param time_max query int false "Max timestamp (seconds)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /boxes/{id}/records [get]
func (h *SensorHandler) ListRecords(c *gin.Context) {
	boxID := c.Param("id")
	if boxID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	pagination := domain.ParsePaginationParams(c)

	var query domain.QueryRecord

	// Parse time range
	if timeMin := c.Query("time_min"); timeMin != "" {
		if timeMax := c.Query("time_max"); timeMax != "" {
			min, _ := strconv.ParseInt(timeMin, 10, 64)
			max, _ := strconv.ParseInt(timeMax, 10, 64)
			query.Time = []int64{min, max}
		}
	}

	limit := pagination.GetLimit()
	skip := pagination.GetSkip()
	query.Limit = &limit
	query.Skip = &skip

	result, err := h.service.ListRecords(c.Request.Context(), boxID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filterInfo := map[string]interface{}{}
	if len(query.Time) == 2 {
		filterInfo["time_min"] = query.Time[0]
		filterInfo["time_max"] = query.Time[1]
	}

	c.JSON(http.StatusOK, domain.NewPaginatedResponse(result.Records, pagination.Page, pagination.PageSize, result.Total, filterInfo))
}

// CountRecords godoc
// @Summary Count sensor records for a box
// @Tags data
// @Security BearerAuth
// @Produce json
// @Param box_id path string true "Box ID"
// @Param time_min query int false "Min timestamp (seconds)"
// @Param time_max query int false "Max timestamp (seconds)"
// @Success 200 {object} map[string]interface{}
// @Router /data/box/{box_id}/count [get]
func (h *SensorHandler) CountRecords(c *gin.Context) {
	boxID := c.Param("box_id")

	var query domain.QueryRecord

	// Parse time range
	if timeMin := c.Query("time_min"); timeMin != "" {
		if timeMax := c.Query("time_max"); timeMax != "" {
			min, _ := strconv.ParseInt(timeMin, 10, 64)
			max, _ := strconv.ParseInt(timeMax, 10, 64)
			query.Time = []int64{min, max}
		}
	}

	count, err := h.service.CountRecords(c.Request.Context(), boxID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// AddRecord godoc
// @Summary Add a sensor record
// @Tags boxes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Box ID"
// @Param request body domain.Record true "Record data"
// @Success 201 {object} map[string]interface{}
// @Router /boxes/{id}/records [post]
func (h *SensorHandler) AddRecord(c *gin.Context) {
	boxID := c.Param("id")

	var record domain.Record
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use ImportRecord service method (assuming it handles the logic)
	if err := h.service.ImportRecord(c.Request.Context(), boxID, record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "record added successfully"})
}

// ReportRecords godoc
// @Summary Generate daily report for a box
// @Tags boxes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Box ID"
// @Param time_min query int false "Min timestamp (seconds)"
// @Param time_max query int false "Max timestamp (seconds)"
// @Success 200 {array} domain.DailyReport
// @Router /boxes/{id}/reports [get]
func (h *SensorHandler) ReportRecords(c *gin.Context) {
	boxID := c.Param("id")
	if boxID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var query domain.QueryRecord

	// Parse time range
	if timeMin := c.Query("time_min"); timeMin != "" {
		if timeMax := c.Query("time_max"); timeMax != "" {
			min, _ := strconv.ParseInt(timeMin, 10, 64)
			max, _ := strconv.ParseInt(timeMax, 10, 64)
			query.Time = []int64{min, max}
		}
	}

	reports, err := h.service.ReportRecords(c.Request.Context(), boxID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// ListRecordsByGroup godoc
// @Summary List sensor records for all boxes in a group
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID"
// @Param time_min query int false "Min timestamp (seconds)"
// @Param time_max query int false "Max timestamp (seconds)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /groups/{id}/records [get]
func (h *SensorHandler) ListRecordsByGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	pagination := domain.ParsePaginationParams(c)

	var query domain.QueryRecord

	if timeMin := c.Query("time_min"); timeMin != "" {
		if timeMax := c.Query("time_max"); timeMax != "" {
			min, _ := strconv.ParseInt(timeMin, 10, 64)
			max, _ := strconv.ParseInt(timeMax, 10, 64)
			query.Time = []int64{min, max}
		}
	}

	limit := pagination.GetLimit()
	skip := pagination.GetSkip()
	query.Limit = &limit
	query.Skip = &skip

	result, err := h.service.ListRecordsByGroup(c.Request.Context(), groupID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filterInfo := map[string]interface{}{}
	if len(query.Time) == 2 {
		filterInfo["time_min"] = query.Time[0]
		filterInfo["time_max"] = query.Time[1]
	}

	c.JSON(http.StatusOK, domain.NewPaginatedResponse(result.Records, pagination.Page, pagination.PageSize, result.Total, filterInfo))
}

// ListRecordsLatestByGroup godoc
// @Summary List sensor records latest for all boxes in a group
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} domain.PaginatedResponse
// @Router /groups/{id}/records/latest [get]
func (h *SensorHandler) ListRecordsLatestByGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	result, err := h.service.ListRecordsLatestByGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.PaginatedResponse{
		Data: result.Records,
		Meta: domain.PaginationMeta{
			TotalItems: result.Total,
		},
	})
}
