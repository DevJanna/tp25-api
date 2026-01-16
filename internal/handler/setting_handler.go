package handler

import (
	"net/http"

	"tp25-api/internal/domain"
	"tp25-api/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type SettingHandler struct {
	service *service.SettingService
}

func NewSettingHandler(service *service.SettingService) *SettingHandler {
	return &SettingHandler{
		service: service,
	}
}

// ListSettings godoc
// @Summary List all settings
// @Tags settings
// @Security BearerAuth
// @Produce json
// @Param key query string false "Filter by key"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} domain.PaginatedResponse
// @Router /settings [get]
func (h *SettingHandler) ListSettings(c *gin.Context) {
	pagination := domain.ParsePaginationParams(c)

	// Build filter
	filter := bson.M{}
	key := c.Query("key")
	if key != "" {
		filter["key"] = key
	}

	settings, total, err := h.service.ListWithPagination(c.Request.Context(), pagination, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Include filter info in response
	var filterInfo interface{}
	if key != "" {
		filterInfo = map[string]string{"key": key}
	}

	response := domain.NewPaginatedResponse(settings, pagination.Page, pagination.PageSize, total, filterInfo)
	c.JSON(http.StatusOK, response)
}

// GetSetting godoc
// @Summary Get setting by ID
// @Tags settings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} domain.Setting
// @Failure 404 {object} map[string]interface{}
// @Router /settings/{id} [get]
func (h *SettingHandler) GetSetting(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	setting, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrSettingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// GetSettingByKey - Deprecated, use GET /settings?key={key} instead
func (h *SettingHandler) GetSettingByKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	setting, err := h.service.GetByKey(c.Request.Context(), key)
	if err != nil {
		if err == domain.ErrSettingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// CreateSetting godoc
// @Summary Create a new setting
// @Tags settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateSettingParams true "Setting data"
// @Success 201 {object} domain.Setting
// @Failure 400 {object} map[string]interface{}
// @Router /settings [post]
func (h *SettingHandler) CreateSetting(c *gin.Context) {
	var params domain.CreateSettingParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.service.Create(c.Request.Context(), params)
	if err != nil {
		if err == domain.ErrSettingKeyExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "setting key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, setting)
}

// UpdateSetting godoc
// @Summary Update a setting
// @Tags settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param request body domain.UpdateSettingParams true "Update data"
// @Success 200 {object} domain.Setting
// @Failure 404 {object} map[string]interface{}
// @Router /settings/{id} [put]
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	id := c.Param("id")

	// Check if updating by key via query param
	key := c.Query("key")
	if key != "" {
		var params domain.UpdateSettingParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		setting, err := h.service.UpdateByKey(c.Request.Context(), key, params)
		if err != nil {
			if err == domain.ErrSettingNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, setting)
		return
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var params domain.UpdateSettingParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.service.Update(c.Request.Context(), id, params)
	if err != nil {
		if err == domain.ErrSettingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateSettingByKey - Deprecated, use PUT /settings/{id}?key={key} instead
func (h *SettingHandler) UpdateSettingByKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	var params domain.UpdateSettingParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.service.UpdateByKey(c.Request.Context(), key, params)
	if err != nil {
		if err == domain.ErrSettingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// DeleteSetting godoc
// @Summary Delete a setting
// @Tags settings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /settings/{id} [delete]
func (h *SettingHandler) DeleteSetting(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrSettingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "setting deleted successfully"})
}
