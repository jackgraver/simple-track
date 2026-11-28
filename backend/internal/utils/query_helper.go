package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QueryParams represents parsed query parameters from HTTP requests
type QueryParams struct {
	Page       int
	PageSize   int
	OrderBy    string
	OrderDesc  bool
	Filters    map[string]string // Field -> value mappings
	ExcludeIDs []uint
	Preloads   []string // Relations to preload, from QueryPolicy.AllowedPreloads
}

type QueryPolicy struct {
	SortableFields   map[string]string
	FilterableFields map[string]string
	AllowedPreloads  map[string]string
}

func ParseQueryParams(c *gin.Context, policy QueryPolicy) (QueryParams, error) {
	params := QueryParams{
		Filters: make(map[string]string),
	}

	// Parse pagination
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "0")
	params.Page, _ = strconv.Atoi(pageStr)
	params.PageSize, _ = strconv.Atoi(pageSizeStr)

	if orderBy := c.Query("orderBy"); orderBy != "" {
		field, ok := policy.SortableFields[orderBy]
		if !ok {
			return QueryParams{}, fmt.Errorf("unsupported sort field %q", orderBy)
		}
		params.OrderBy = field
	}
	orderDescStr := c.DefaultQuery("orderDesc", "false")
	params.OrderDesc = orderDescStr == "true"

	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "pageSize" && key != "orderBy" && key != "orderDesc" &&
			key != "exclude" && key != "preloads" && key != "useDefaultPreloads" {
			if len(values) > 0 {
				field, ok := policy.FilterableFields[key]
				if !ok {
					return QueryParams{}, fmt.Errorf("unsupported filter field %q", key)
				}
				params.Filters[field] = values[0]
			}
		}
	}

	// Parse exclude IDs
	if excludeStr := c.Query("exclude"); excludeStr != "" {
		ids := []uint{}
		for _, idStr := range []string{excludeStr} {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
		params.ExcludeIDs = ids
	}

	if preloadsStr := c.Query("preloads"); preloadsStr != "" {
		for _, requested := range strings.Split(preloadsStr, ",") {
			requested = strings.TrimSpace(requested)
			preload, ok := policy.AllowedPreloads[requested]
			if !ok {
				return QueryParams{}, fmt.Errorf("unsupported preload %q", requested)
			}
			params.Preloads = append(params.Preloads, preload)
		}
	}

	return params, nil
}

// Pagination describes one page of a list query.
type Pagination struct {
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

// ListResult is the rows of a list query, with Pagination set when a page was requested.
type ListResult[T any] struct {
	Data       []T
	Pagination *Pagination
}

// List queries T with params' filters, excluded IDs, preloads and ordering (defaultOrderBy
// when params has none). It returns one page when params.Page and params.PageSize are both set,
// otherwise every matching row. Field names in params come from a QueryPolicy allowlist.
func List[T any](db *gorm.DB, params QueryParams, defaultOrderBy string, defaultOrderDesc bool) (*ListResult[T], error) {
	q := db.Model(new(T))
	for field, value := range params.Filters {
		q = q.Where(clause.Eq{Column: clause.Column{Name: field}, Value: value})
	}
	if len(params.ExcludeIDs) > 0 {
		q = q.Where("id NOT IN ?", params.ExcludeIDs)
	}

	var page *Pagination
	if params.Page > 0 && params.PageSize > 0 {
		var total int64
		if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return nil, err
		}
		totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
		page = &Pagination{
			Total:      total,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		}
		q = q.Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize)
	}

	for _, preload := range params.Preloads {
		q = q.Preload(preload)
	}
	orderBy, orderDesc := params.OrderBy, params.OrderDesc
	if orderBy == "" {
		orderBy, orderDesc = defaultOrderBy, defaultOrderDesc
	}
	if orderBy != "" {
		q = q.Order(clause.OrderByColumn{Column: clause.Column{Name: orderBy}, Desc: orderDesc})
	}

	var rows []T
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return &ListResult[T]{Data: rows, Pagination: page}, nil
}
