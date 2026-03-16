package utils

import (
	"math"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Pagination holds pagination parameters
type Pagination struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
	Total    int64 `json:"total"`
}

// PaginatedResult holds paginated results
type PaginatedResult struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPagination creates a new pagination instance
func NewPagination(page, pageSize int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return &Pagination{Page: page, PageSize: pageSize}
}

// Offset returns the offset for database queries
func (p *Pagination) Offset() int64 {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the limit for database queries
func (p *Pagination) Limit() int64 {
	return p.PageSize
}

// SetTotal sets the total count
func (p *Pagination) SetTotal(total int64) {
	p.Total = total
}

// GetMeta returns pagination metadata
func (p *Pagination) GetMeta() PaginationMeta {
	totalPages := int64(math.Ceil(float64(p.Total) / float64(p.PageSize)))
	if totalPages < 1 {
		totalPages = 1
	}
	return PaginationMeta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      p.Total,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}

// NewPaginatedResult creates a paginated result
func NewPaginatedResult(data interface{}, pagination *Pagination) *PaginatedResult {
	return &PaginatedResult{
		Data:       data,
		Pagination: pagination.GetMeta(),
	}
}
