package repository

import "time"

// QueryOptions configures how a repository query is executed
type QueryOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Preloading - nil means use entity defaults, empty slice means no preloads
	Preloads    []string
	UseDefaults bool // If true, use entity's default preloads

	// Sorting
	OrderBy   string
	OrderDesc bool

	// Date range filtering (for Dateable entities)
	DateStart *time.Time
	DateEnd   *time.Time

	// Generic filters (field -> value)
	Filters map[string]interface{}

	// Exclude IDs
	ExcludeIDs []uint
}

// QueryOption is a functional option for configuring queries
type QueryOption func(*QueryOptions)

// DefaultOptions returns QueryOptions with sensible defaults
func DefaultOptions() *QueryOptions {
	return &QueryOptions{
		Limit:       0, // 0 means no limit
		Offset:      0,
		Preloads:    nil,
		UseDefaults: true, // Use entity's default preloads
		OrderBy:     "",
		OrderDesc:   false,
		Filters:     make(map[string]interface{}),
		ExcludeIDs:  nil,
	}
}

// ApplyOptions applies functional options to QueryOptions
func ApplyOptions(opts ...QueryOption) *QueryOptions {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithPagination sets limit and offset for pagination
func WithPagination(page, pageSize int) QueryOption {
	return func(o *QueryOptions) {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
		o.Limit = pageSize
		o.Offset = (page - 1) * pageSize
	}
}

// WithLimit sets a limit on results
func WithLimit(limit int) QueryOption {
	return func(o *QueryOptions) {
		o.Limit = limit
	}
}

// WithOffset sets an offset for results
func WithOffset(offset int) QueryOption {
	return func(o *QueryOptions) {
		o.Offset = offset
	}
}

// WithPreloads specifies which relationships to eager load
// Pass nil or empty to disable preloading, pass specific paths to override defaults
func WithPreloads(preloads ...string) QueryOption {
	return func(o *QueryOptions) {
		o.Preloads = preloads
		o.UseDefaults = false
	}
}

// WithDefaultPreloads uses the entity's default preloads
func WithDefaultPreloads() QueryOption {
	return func(o *QueryOptions) {
		o.UseDefaults = true
		o.Preloads = nil
	}
}

// WithNoPreloads disables all preloading
func WithNoPreloads() QueryOption {
	return func(o *QueryOptions) {
		o.UseDefaults = false
		o.Preloads = []string{}
	}
}

// WithOrderBy sets the ordering field and direction
func WithOrderBy(field string, desc bool) QueryOption {
	return func(o *QueryOptions) {
		o.OrderBy = field
		o.OrderDesc = desc
	}
}

// WithOrderByAsc sets ascending order on a field
func WithOrderByAsc(field string) QueryOption {
	return WithOrderBy(field, false)
}

// WithOrderByDesc sets descending order on a field
func WithOrderByDesc(field string) QueryOption {
	return WithOrderBy(field, true)
}

// WithDateRange filters by date range (for Dateable entities)
func WithDateRange(start, end time.Time) QueryOption {
	return func(o *QueryOptions) {
		o.DateStart = &start
		o.DateEnd = &end
	}
}

// WithDateFrom filters from a start date
func WithDateFrom(start time.Time) QueryOption {
	return func(o *QueryOptions) {
		o.DateStart = &start
	}
}

// WithDateUntil filters until an end date
func WithDateUntil(end time.Time) QueryOption {
	return func(o *QueryOptions) {
		o.DateEnd = &end
	}
}

// WithFilter adds a single filter condition
func WithFilter(field string, value interface{}) QueryOption {
	return func(o *QueryOptions) {
		if o.Filters == nil {
			o.Filters = make(map[string]interface{})
		}
		o.Filters[field] = value
	}
}

// WithFilters adds multiple filter conditions
func WithFilters(filters map[string]interface{}) QueryOption {
	return func(o *QueryOptions) {
		for k, v := range filters {
			if o.Filters == nil {
				o.Filters = make(map[string]interface{})
			}
			o.Filters[k] = v
		}
	}
}

// WithExcludeIDs excludes specific IDs from results
func WithExcludeIDs(ids ...uint) QueryOption {
	return func(o *QueryOptions) {
		o.ExcludeIDs = ids
	}
}

// PaginatedResult wraps query results with pagination metadata
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginatedResult creates a PaginatedResult from data and counts
func NewPaginatedResult[T any](data []T, total int64, page, pageSize int) *PaginatedResult[T] {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
