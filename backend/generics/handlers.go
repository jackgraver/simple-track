package generics

import (
	"be-simpletracker/database/repository"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CRUDConfig holds configuration for CRUD operations
type CRUDConfig[T repository.Entity] struct {
	// Route configuration
	BasePath     string // e.g., "/plans" or "/exercises"
	ResourceName string // e.g., "plan" or "exercise" (for JSON keys)

	// Route flags - control which routes are registered
	EnableGetAll  bool // Default: true - enables GET /base/all
	EnableGetByID bool // Default: true - enables GET /base/:id
	EnableCreate  bool // Default: true - enables POST /base
	EnableUpdate  bool // Default: true - enables PUT /base/:id
	EnableDelete  bool // Default: true - enables DELETE /base/:id

	// Default query options
	DefaultPageSize  int    // Default: 20
	DefaultOrderBy   string // Default: "id"
	DefaultOrderDesc bool   // Default: true

	// Preloads
	UseDefaultPreloads bool // Default: true

	// Custom handlers (optional - for entity-specific logic)
	BeforeCreate func(ctx context.Context, db *gorm.DB, entity *T) error
	AfterCreate  func(ctx context.Context, db *gorm.DB, entity *T) error
	BeforeUpdate func(ctx context.Context, db *gorm.DB, entity *T) error
	AfterUpdate  func(ctx context.Context, db *gorm.DB, entity *T) error
	BeforeDelete func(ctx context.Context, db *gorm.DB, id uint) error
	AfterDelete  func(ctx context.Context, db *gorm.DB, id uint) error

	// Custom query options builder
	BuildQueryOptions func(c *gin.Context) []repository.QueryOption
}

// GenericCRUDHandler holds the database connection and config
type GenericCRUDHandler[T repository.Entity] struct {
	db     *gorm.DB
	config CRUDConfig[T]
}

// DefaultCRUDConfig creates a default configuration with sensible defaults
// All routes are enabled by default
func DefaultCRUDConfig[T repository.Entity](basePath, resourceName string) CRUDConfig[T] {
	return CRUDConfig[T]{
		BasePath:            basePath,
		ResourceName:        resourceName,
		EnableGetAll:        true,
		EnableGetByID:       true,
		EnableCreate:        true,
		EnableUpdate:        true,
		EnableDelete:        true,
		DefaultPageSize:     20,
		DefaultOrderBy:      "id",
		DefaultOrderDesc:    true,
		UseDefaultPreloads:  true,
	}
}

// RegisterBasicCRUD registers standard CRUD routes for an entity
// Only registers routes that are enabled in the config
func RegisterBasicCRUD[T repository.Entity](
	group *gin.RouterGroup,
	db *gorm.DB,
	config CRUDConfig[T],
) {
	handler := &GenericCRUDHandler[T]{
		db:     db,
		config: config,
	}

	base := group.Group(config.BasePath)
	{
		if config.EnableGetAll {
			base.GET("/all", handler.GetAll) // GET /base/all
		}
		if config.EnableGetByID {
			base.GET("/:id", handler.GetByID) // GET /base/:id
		}
		if config.EnableCreate {
			base.POST("", handler.Create) // POST /base
		}
		if config.EnableUpdate {
			base.PUT("/:id", handler.Update) // PUT /base/:id
		}
		if config.EnableDelete {
			base.DELETE("/:id", handler.Delete) // DELETE /base/:id
		}
	}
}

// GetAll retrieves all entities with support for pagination, sorting, and filtering
func (h *GenericCRUDHandler[T]) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "0")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// Build query options
	var opts []repository.QueryOption

	// Use custom query options builder if provided
	if h.config.BuildQueryOptions != nil {
		opts = h.config.BuildQueryOptions(c)
	} else {
		// Build default query options
		opts = h.buildDefaultQueryOptions(c)
	}

	// Execute query with pagination or without
	if page > 0 && pageSize > 0 {
		// Paginated query
		result, err := GetAllPaginated[T](ctx, h.db, page, pageSize, opts...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		responseKey := h.config.ResourceName + "s"
		c.JSON(http.StatusOK, gin.H{
			responseKey: result.Data,
			"pagination": gin.H{
				"total":      result.Total,
				"page":       result.Page,
				"pageSize":   result.PageSize,
				"totalPages": result.TotalPages,
				"hasNext":    result.HasNext,
				"hasPrev":    result.HasPrev,
			},
		})
	} else {
		// Non-paginated query (use default page size if configured)
		if h.config.DefaultPageSize > 0 {
			result, err := GetAllPaginated[T](ctx, h.db, 1, h.config.DefaultPageSize, opts...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			responseKey := h.config.ResourceName + "s"
			c.JSON(http.StatusOK, gin.H{
				responseKey: result.Data,
				"pagination": gin.H{
					"total":      result.Total,
					"page":       result.Page,
					"pageSize":   result.PageSize,
					"totalPages": result.TotalPages,
					"hasNext":    result.HasNext,
					"hasPrev":    result.HasPrev,
				},
			})
		} else {
			// No pagination
			entities, err := GetAll[T](ctx, h.db, opts...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			responseKey := h.config.ResourceName + "s"
			c.JSON(http.StatusOK, gin.H{
				responseKey: entities,
			})
		}
	}
}

// GetByID retrieves a single entity by ID
func (h *GenericCRUDHandler[T]) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	id := uint(idUint64)

	// Build query options
	var opts []repository.QueryOption
	if h.config.UseDefaultPreloads {
		opts = append(opts, repository.WithDefaultPreloads())
	}

	// Use generic GetOne
	entity, err := GetOne[T](ctx, h.db, id, opts...)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", h.config.ResourceName)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{h.config.ResourceName: entity})
}

// Create creates a new entity
func (h *GenericCRUDHandler[T]) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var entity T
	if err := c.BindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call BeforeCreate hook if provided
	if h.config.BeforeCreate != nil {
		if err := h.config.BeforeCreate(ctx, h.db, &entity); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Use generic Create
	if err := Create(ctx, h.db, &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Call AfterCreate hook if provided
	if h.config.AfterCreate != nil {
		if err := h.config.AfterCreate(ctx, h.db, &entity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{h.config.ResourceName: entity})
}

// Update updates an existing entity
func (h *GenericCRUDHandler[T]) Update(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	id := uint(idUint64)

	var entity T
	if err := c.BindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure ID matches URL parameter using reflection
	// Since we can't directly set ID on a generic type, we use reflection
	entityValue := reflect.ValueOf(&entity).Elem()
	idField := entityValue.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		// Check if ID in entity matches URL parameter
		currentID := uint(idField.Uint())
		if currentID != 0 && currentID != id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID in body does not match URL parameter"})
			return
		}
		// Set ID from URL parameter
		idField.SetUint(uint64(id))
	} else {
		// If ID field is not accessible, validate that GetID() matches
		if entity.GetID() != 0 && entity.GetID() != id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID in body does not match URL parameter"})
			return
		}
		// Try to set via Model field (gorm.Model)
		modelField := entityValue.FieldByName("Model")
		if modelField.IsValid() {
			modelValue := modelField
			if modelValue.Kind() == reflect.Struct {
				modelIDField := modelValue.FieldByName("ID")
				if modelIDField.IsValid() && modelIDField.CanSet() {
					modelIDField.SetUint(uint64(id))
				}
			}
		}
	}

	// Call BeforeUpdate hook if provided
	if h.config.BeforeUpdate != nil {
		if err := h.config.BeforeUpdate(ctx, h.db, &entity); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Use generic Update
	if err := Update(ctx, h.db, &entity); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", h.config.ResourceName)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Call AfterUpdate hook if provided
	if h.config.AfterUpdate != nil {
		if err := h.config.AfterUpdate(ctx, h.db, &entity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{h.config.ResourceName: entity})
}

// Delete deletes an entity by ID (soft delete)
func (h *GenericCRUDHandler[T]) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	id := uint(idUint64)

	// Call BeforeDelete hook if provided
	if h.config.BeforeDelete != nil {
		if err := h.config.BeforeDelete(ctx, h.db, id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Use generic Delete (always soft delete)
	if err := Delete[T](ctx, h.db, id); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", h.config.ResourceName)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Call AfterDelete hook if provided
	if h.config.AfterDelete != nil {
		if err := h.config.AfterDelete(ctx, h.db, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// buildDefaultQueryOptions builds query options from request parameters
func (h *GenericCRUDHandler[T]) buildDefaultQueryOptions(c *gin.Context) []repository.QueryOption {
	var opts []repository.QueryOption

	// Add preloads
	if h.config.UseDefaultPreloads {
		opts = append(opts, repository.WithDefaultPreloads())
	}

	// Add sorting
	orderBy := c.Query("orderBy")
	if orderBy != "" {
		orderDesc := c.DefaultQuery("orderDesc", "false") == "true"
		opts = append(opts, repository.WithOrderBy(orderBy, orderDesc))
	} else if h.config.DefaultOrderBy != "" {
		opts = append(opts, repository.WithOrderBy(h.config.DefaultOrderBy, h.config.DefaultOrderDesc))
	}

	// Add filters (any query param that's not reserved)
	reservedParams := map[string]bool{
		"page":       true,
		"pageSize":   true,
		"orderBy":    true,
		"orderDesc":  true,
		"exclude":    true,
		"preloads":   true,
		"useDefaultPreloads": true,
	}

	for key, values := range c.Request.URL.Query() {
		if !reservedParams[key] && len(values) > 0 {
			opts = append(opts, repository.WithFilter(key, values[0]))
		}
	}

	// Add exclude IDs if provided
	if excludeStr := c.Query("exclude"); excludeStr != "" {
		ids := []uint{}
		for _, idStr := range []string{excludeStr} {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) > 0 {
			opts = append(opts, repository.WithExcludeIDs(ids...))
		}
	}

	return opts
}
