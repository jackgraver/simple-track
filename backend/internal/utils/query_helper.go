package utils

import (
	"be-simpletracker/internal/database/repository"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QueryParams represents parsed query parameters from HTTP requests
type QueryParams struct {
	Page               int
	PageSize           int
	OrderBy            string
	OrderDesc          bool
	Filters            map[string]string // Field -> value mappings
	ExcludeIDs         []uint
	Preloads           []string // Custom preloads (empty means use defaults, nil means no preloads)
	UseDefaultPreloads bool
}

type QueryPolicy struct {
	SortableFields   map[string]string
	FilterableFields map[string]string
	AllowedPreloads  map[string]string
}

func ParseQueryParams(c *gin.Context, policy QueryPolicy) (QueryParams, error) {
	params := QueryParams{
		Filters:            make(map[string]string),
		UseDefaultPreloads: true,
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
		params.UseDefaultPreloads = false
	}

	if c.Query("useDefaultPreloads") == "false" {
		params.UseDefaultPreloads = false
	}

	return params, nil
}

// BuildQueryOptions converts QueryParams to repository.QueryOption slice
func BuildQueryOptions(params QueryParams, defaultOrderBy string, defaultOrderDesc bool) []repository.QueryOption {
	var opts []repository.QueryOption

	// Add pagination
	if params.Page > 0 && params.PageSize > 0 {
		opts = append(opts, repository.WithPagination(params.Page, params.PageSize))
	}

	// Add sorting
	if params.OrderBy != "" {
		opts = append(opts, repository.WithOrderBy(params.OrderBy, params.OrderDesc))
	} else if defaultOrderBy != "" {
		opts = append(opts, repository.WithOrderBy(defaultOrderBy, defaultOrderDesc))
	}

	// Add filters
	for field, value := range params.Filters {
		opts = append(opts, repository.WithFilter(field, value))
	}

	// Add exclude IDs
	if len(params.ExcludeIDs) > 0 {
		opts = append(opts, repository.WithExcludeIDs(params.ExcludeIDs...))
	}

	// Add preloads
	if params.UseDefaultPreloads {
		opts = append(opts, repository.WithDefaultPreloads())
	} else if len(params.Preloads) > 0 {
		opts = append(opts, repository.WithPreloads(params.Preloads...))
	} else {
		opts = append(opts, repository.WithNoPreloads())
	}

	return opts
}

// GetAllResult represents the result of a GetAll query (either paginated or not)
type GetAllResult[T repository.Entity] struct {
	Data       []T
	Pagination *repository.PaginatedResult[T]
}

// GetAllWithOptions is a convenience function that handles the full flow:
// 1. Parse query params from gin.Context
// 2. Build query options
// 3. Execute query (paginated or not)
// 4. Return results
func GetAllWithOptions[T repository.Entity](
	ctx context.Context,
	db *gorm.DB,
	c *gin.Context,
	defaultOrderBy string,
	defaultOrderDesc bool,
	policy QueryPolicy,
) (*GetAllResult[T], error) {
	repo := repository.NewGormRepository[T](db)

	params, err := ParseQueryParams(c, policy)
	if err != nil {
		return nil, err
	}
	opts := BuildQueryOptions(params, defaultOrderBy, defaultOrderDesc)

	// Use paginated query if pagination params are provided
	if params.Page > 0 && params.PageSize > 0 {
		result, err := repo.GetAllPaginated(ctx, params.Page, params.PageSize, opts...)
		if err != nil {
			return nil, err
		}
		return &GetAllResult[T]{
			Data:       result.Data,
			Pagination: result,
		}, nil
	}

	// Use non-paginated query
	entities, err := repo.GetAll(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &GetAllResult[T]{
		Data:       entities,
		Pagination: nil,
	}, nil
}
