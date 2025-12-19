package services

import (
	"be-simpletracker/database/repository"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QueryParams represents parsed query parameters from HTTP requests
type QueryParams struct {
	Page      int
	PageSize  int
	OrderBy   string
	OrderDesc bool
	Filters   map[string]string // Field -> value mappings
	ExcludeIDs []uint
	Preloads  []string // Custom preloads (empty means use defaults, nil means no preloads)
	UseDefaultPreloads bool
}

// ParseQueryParams extracts query parameters from gin.Context
func ParseQueryParams(c *gin.Context) QueryParams {
	params := QueryParams{
		Filters: make(map[string]string),
		UseDefaultPreloads: true,
	}

	// Parse pagination
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "0")
	params.Page, _ = strconv.Atoi(pageStr)
	params.PageSize, _ = strconv.Atoi(pageSizeStr)

	// Parse sorting
	params.OrderBy = c.Query("orderBy")
	orderDescStr := c.DefaultQuery("orderDesc", "false")
	params.OrderDesc = orderDescStr == "true"

	// Parse filters - support multiple filter formats
	// Format 1: ?name=value
	// Format 2: ?filter=field:value (for future expansion)
	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "pageSize" && key != "orderBy" && key != "orderDesc" && 
		   key != "exclude" && key != "preloads" && key != "useDefaultPreloads" {
			// Treat other query params as filters
			if len(values) > 0 {
				params.Filters[key] = values[0]
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

	// Parse preloads
	if preloadsStr := c.Query("preloads"); preloadsStr != "" {
		params.Preloads = []string{preloadsStr} // Could split by comma for multiple
		params.UseDefaultPreloads = false
	}

	if c.Query("useDefaultPreloads") == "false" {
		params.UseDefaultPreloads = false
	}

	return params
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
	// repo repository.Repository[T],
	db *gorm.DB,
	c *gin.Context,
	defaultOrderBy string,
	defaultOrderDesc bool,
) (*GetAllResult[T], error) {
	repo := repository.NewGormRepository[T](db)

	params := ParseQueryParams(c)
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
