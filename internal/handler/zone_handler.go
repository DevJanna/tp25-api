package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"tp25-api/internal/domain"
	"tp25-api/internal/service"
)

type ZoneHandler struct {
	service *service.ZoneService
}

func NewZoneHandler(service *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{service: service}
}

// Zone endpoints

// ListZones godoc
// @Summary List all zones
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /zones [get]
func (h *ZoneHandler) ListZones(c *gin.Context) {
	pagination := domain.ParsePaginationParams(c)

	// Build filter
	filter := bson.M{}

	zones, total, err := h.service.ListZonesWithPagination(c.Request.Context(), pagination, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := domain.NewPaginatedResponse(zones, pagination.Page, pagination.PageSize, total, filter)
	c.JSON(http.StatusOK, response)
}

// GetZone godoc
// @Summary Get zone by ID
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {object} domain.Zone
// @Failure 404 {object} map[string]interface{}
// @Router /zones/{id} [get]
func (h *ZoneHandler) GetZone(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	zone, err := h.service.GetZone(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrZoneNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// CreateZone godoc
// @Summary Create a new zone
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateZoneParams true "Zone data"
// @Success 201 {object} domain.Zone
// @Failure 409 {object} map[string]interface{}
// @Router /zones [post]
func (h *ZoneHandler) CreateZone(c *gin.Context) {
	var params domain.CreateZoneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.service.CreateZone(c.Request.Context(), params)
	if err != nil {
		if err == domain.ErrZoneCodeExisted {
			c.JSON(http.StatusConflict, gin.H{"error": "zone code already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, zone)
}

// UpdateZone godoc
// @Summary Update zone
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Param request body domain.UpdateZoneParams true "Update data"
// @Success 200 {object} domain.Zone
// @Failure 404 {object} map[string]interface{}
// @Router /zones/{id} [put]
func (h *ZoneHandler) UpdateZone(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var params domain.UpdateZoneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.service.UpdateZone(c.Request.Context(), id, params)
	if err != nil {
		if err == domain.ErrZoneNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// BoxGroup endpoints

// ListGroups godoc
// @Summary List box groups
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {array} domain.ViewBox
// @Router /zones/{id}/groups [get]
func (h *ZoneHandler) ListGroups(c *gin.Context) {
	zoneID := c.Param("id")

	groups, err := h.service.ListGroups(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// GetGroup godoc
// @Summary Get box group by ID
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} domain.ViewBox
// @Failure 404 {object} map[string]interface{}
// @Router /groups/{id} [get]
func (h *ZoneHandler) GetGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	group, err := h.service.GetGroup(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrBoxGroupNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// CreateGroup godoc
// @Summary Create a new box group
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Param request body domain.CreateGroupParams true "Group data"
// @Success 201 {object} domain.BoxGroup
// @Router /zones/{id}/groups [post]
func (h *ZoneHandler) CreateGroup(c *gin.Context) {
	var params domain.CreateGroupParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override zone_id from path param to ensure consistency
	zoneID := c.Param("id")
	if zoneID != "" {
		params.ZoneID = zoneID
	}

	group, err := h.service.CreateGroup(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup godoc
// @Summary Update box group
// @Tags groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param request body domain.UpdateGroupParams true "Update data"
// @Success 200 {object} domain.ViewBox
// @Failure 404 {object} map[string]interface{}
// @Router /groups/{id} [put]
func (h *ZoneHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var params domain.UpdateGroupParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.UpdateGroup(c.Request.Context(), id, params)
	if err != nil {
		if err == domain.ErrBoxGroupNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup godoc
// @Summary Delete box group (soft delete)
// @Tags groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} domain.BoxGroup
// @Failure 404 {object} map[string]interface{}
// @Router /groups/{id} [delete]
func (h *ZoneHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	group, err := h.service.DeleteGroup(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrBoxGroupNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// Box endpoints

// ListAllBoxes godoc
// @Summary List all boxes
// @Tags boxes
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /boxes [get]
func (h *ZoneHandler) ListAllBoxes(c *gin.Context) {
	pagination := domain.ParsePaginationParams(c)

	var filter domain.FilterBoxParams

	boxes, total, err := h.service.ListBoxesWithPagination(c.Request.Context(), pagination, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := domain.NewPaginatedResponse(boxes, pagination.Page, pagination.PageSize, total, nil)
	c.JSON(http.StatusOK, response)
}

// ListBoxes godoc
// @Summary List boxes
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /groups/{id}/boxes [get]
func (h *ZoneHandler) ListBoxes(c *gin.Context) {
	groupID := c.Param("id")

	pagination := domain.ParsePaginationParams(c)

	var filter domain.FilterBoxParams
	filter.GroupID = &groupID

	boxes, total, err := h.service.ListBoxesWithPagination(c.Request.Context(), pagination, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build filter info
	filterInfo := map[string]interface{}{
		"group_id": groupID,
	}

	response := domain.NewPaginatedResponse(boxes, pagination.Page, pagination.PageSize, total, filterInfo)
	c.JSON(http.StatusOK, response)
}

// GetBox godoc
// @Summary Get box by ID
// @Tags boxes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Box ID"
// @Success 200 {object} domain.Box
// @Failure 404 {object} map[string]interface{}
// @Router /boxes/{id} [get]
func (h *ZoneHandler) GetBox(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	box, err := h.service.GetBox(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrBoxNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, box)
}

// CreateBox godoc
// @Summary Create a new box
// @Tags groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param request body domain.CreateBoxParams true "Box data"
// @Success 201 {object} domain.Box
// @Failure 409 {object} map[string]interface{}
// @Router /groups/{id}/boxes [post]
func (h *ZoneHandler) CreateBox(c *gin.Context) {
	var params domain.CreateBoxParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override group_id from path param to ensure consistency
	groupID := c.Param("id")
	if groupID != "" {
		params.GroupID = groupID
	}

	box, err := h.service.CreateBox(c.Request.Context(), params)
	if err != nil {
		if err == domain.ErrBoxDeviceExisted {
			c.JSON(http.StatusConflict, gin.H{"error": "box device already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, box)
}

// UpdateBox godoc
// @Summary Update box
// @Tags boxes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Box ID"
// @Param request body domain.UpdateBoxParams true "Update data"
// @Success 200 {object} domain.Box
// @Failure 404 {object} map[string]interface{}
// @Router /boxes/{id} [put]
func (h *ZoneHandler) UpdateBox(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var params domain.UpdateBoxParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	box, err := h.service.UpdateBox(c.Request.Context(), id, params)
	if err != nil {
		if err == domain.ErrBoxNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, box)
}

// DeleteBox godoc
// @Summary Delete box (soft delete)
// @Tags boxes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Box ID"
// @Success 200 {object} domain.Box
// @Failure 404 {object} map[string]interface{}
// @Router /boxes/{id} [delete]
func (h *ZoneHandler) DeleteBox(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	box, err := h.service.DeleteBox(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrBoxNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, box)
}

// Report endpoint

// ReportByMetric godoc
// @Summary Generate report by metrics
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param group query string true "Group ID"
// @Param metrics query string false "Comma-separated metrics list"
// @Success 200 {array} domain.Report
// @Router /zones/reports [get]
func (h *ZoneHandler) ReportByMetric(c *gin.Context) {
	groupID := c.Query("group")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group parameter is required"})
		return
	}

	metricsStr := c.Query("metrics")
	if metricsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metrics parameter is required"})
		return
	}

	// Split comma-separated metrics
	metrics := []string{}
	for _, m := range splitAndTrim(metricsStr, ",") {
		if m != "" {
			metrics = append(metrics, m)
		}
	}

	if len(metrics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one metric is required"})
		return
	}

	reports, err := h.service.ReportByMetric(c.Request.Context(), groupID, metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimString(part)
		parts = append(parts, trimmed)
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" || len(result) > 0 {
		result = append(result, current)
	}
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
