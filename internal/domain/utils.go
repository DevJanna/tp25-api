package domain

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page" form:"page"`           // Current page (1-indexed)
	PageSize int `json:"page_size" form:"page_size"` // Number of items per page
}

type PaginationMeta struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Filter     interface{} `json:"filter,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// NewPagination creates a new Pagination with default values
func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetSkip calculates the number of items to skip
func (p *Pagination) GetSkip() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

func NewPaginatedResponse(data interface{}, page, pageSize int, totalItems int64, filter interface{}) *PaginatedResponse {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		Data: data,
		Meta: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			Filter:     filter,
		},
	}
}

// ParsePaginationParams parses pagination parameters from gin.Context
func ParsePaginationParams(c *gin.Context) *Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	return NewPagination(page, pageSize)
}
