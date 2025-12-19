# Generic CRUD Handlers Implementation Plan

## Overview
Create a generic CRUD handler system that allows any entity to get full CRUD endpoints with a single function call, eliminating repetitive handler code.

## Goals
- **Simplify API files**: Replace 150+ lines of repetitive handler code with a single function call
- **Consistent behavior**: All entities get the same CRUD operations with consistent error handling
- **Configurable defaults**: Allow customization of pagination, preloads, sorting, etc.
- **Extensible**: Support custom handlers for entity-specific operations

## Architecture

### 1. Create Generic Handler Package
**Location**: `backend/generics/handlers.go` (or `backend/http/handlers.go`)

#### Core Components:

```go
// CRUDConfig holds configuration for CRUD operations
type CRUDConfig[T repository.Entity] struct {
    // Route configuration
    BasePath      string  // e.g., "/plans" or "/exercises"
    ResourceName  string  // e.g., "plan" or "exercise" (for JSON keys)
    
    // Default query options
    DefaultPageSize   int    // Default: 20
    DefaultOrderBy    string // Default: "id"
    DefaultOrderDesc  bool   // Default: true
    
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
```

#### Generic Handler Functions:

```go
// GenericCRUDHandler holds the database connection and config
type GenericCRUDHandler[T repository.Entity] struct {
    db     *gorm.DB
    config CRUDConfig[T]
}

// Handler functions (internal, used by route registration)
func (h *GenericCRUDHandler[T]) GetAll(c *gin.Context)
func (h *GenericCRUDHandler[T]) GetByID(c *gin.Context)
func (h *GenericCRUDHandler[T]) Create(c *gin.Context)
func (h *GenericCRUDHandler[T]) Update(c *gin.Context)
func (h *GenericCRUDHandler[T]) Delete(c *gin.Context)
```

### 2. Route Registration Function

```go
// RegisterBasicCRUD registers standard CRUD routes for an entity
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
        base.GET("/all", handler.GetAll)      // GET /base/all
        base.GET("/:id", handler.GetByID)     // GET /base/:id
        base.POST("", handler.Create)         // POST /base
        base.PUT("/:id", handler.Update)      // PUT /base/:id
        base.DELETE("/:id", handler.Delete)   // DELETE /base/:id
    }
}
```

### 3. Default Configuration Helper

```go
// DefaultCRUDConfig creates a default configuration with sensible defaults
func DefaultCRUDConfig[T repository.Entity](basePath, resourceName string) CRUDConfig[T] {
    return CRUDConfig[T]{
        BasePath:            basePath,
        ResourceName:        resourceName,
        DefaultPageSize:     20,
        DefaultOrderBy:      "id",
        DefaultOrderDesc:    true,
        UseDefaultPreloads:  true,
    }
}
```

## Implementation Steps

### Step 1: Create Generic Handler Package
**File**: `backend/generics/handlers.go`

**Features to implement:**
- Generic handler struct with config
- All 5 CRUD handler methods
- Query parameter parsing (pagination, sorting, filtering)
- Consistent error handling
- JSON response formatting

**Key implementation details:**
- `GetAll`: Parse pagination params, use default page size if not provided, support sorting/filtering
- `GetByID`: Parse ID from URL, handle NotFound errors
- `Create`: Bind JSON, call BeforeCreate hook, create entity, call AfterCreate hook
- `Update`: Parse ID, bind JSON, set ID, call BeforeUpdate hook, update, call AfterUpdate hook
- `Delete`: Parse ID, call BeforeDelete hook, soft delete, call AfterDelete hook

### Step 2: Query Parameter Parsing
**Helper functions:**
- Parse pagination (page, pageSize)
- Parse sorting (orderBy, orderDesc)
- Parse filters (any query param that's not reserved)
- Build QueryOptions from parsed params

### Step 3: Response Formatting
**Standardized responses:**
- `GetAll`: `{"plans": [...], "pagination": {...}}` or `{"plans": [...]}`
- `GetByID`: `{"plan": {...}}`
- `Create`: `{"plan": {...}}` (201 Created)
- `Update`: `{"plan": {...}}` (200 OK)
- `Delete`: `{"success": true}` (200 OK)

### Step 4: Error Handling
**Standard error responses:**
- Invalid ID: `400 Bad Request`
- Not Found: `404 Not Found`
- Validation errors: `400 Bad Request`
- Server errors: `500 Internal Server Error`

### Step 5: Update API Files
**Before (150+ lines):**
```go
func RegisterWorkoutPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
    h := NewWorkoutPlanHandler(db)
    plans := group.Group("/plans")
    {
        plans.GET("/all", h.GetAll)
        plans.GET("/:id", h.GetByID)
        plans.POST("", h.Create)
        plans.PUT("/:id", h.Update)
        plans.DELETE("/:id", h.Delete)
    }
}
// ... 150+ lines of handler methods
```

**After (5-10 lines):**
```go
func RegisterWorkoutPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
    config := generics.DefaultCRUDConfig[models.WorkoutPlan]("/plans", "plan")
    // Optional: customize config
    config.DefaultPageSize = 10
    config.DefaultOrderBy = "name"
    
    generics.RegisterBasicCRUD(group, db, config)
}
```

## Configuration Examples

### Example 1: Basic Usage (WorkoutPlan)
```go
config := generics.DefaultCRUDConfig[models.WorkoutPlan]("/plans", "plan")
generics.RegisterBasicCRUD(group, db, config)
```

### Example 2: Custom Pagination (Exercise)
```go
config := generics.DefaultCRUDConfig[models.Exercise]("/exercises", "exercise")
config.DefaultPageSize = 50
config.DefaultOrderBy = "name"
config.DefaultOrderDesc = false
generics.RegisterBasicCRUD(group, db, config)
```

### Example 3: Custom Preloads (WorkoutLog)
```go
config := generics.DefaultCRUDConfig[models.WorkoutLog]("/logs", "log")
config.BuildQueryOptions = func(c *gin.Context) []repository.QueryOption {
    opts := []repository.QueryOption{
        repository.WithDefaultPreloads(),
    }
    // Add date filter if provided
    if dateStr := c.Query("date"); dateStr != "" {
        // Custom date filtering logic
    }
    return opts
}
generics.RegisterBasicCRUD(group, db, config)
```

### Example 4: With Hooks (for validation/business logic)
```go
config := generics.DefaultCRUDConfig[models.WorkoutPlan]("/plans", "plan")
config.BeforeCreate = func(ctx context.Context, db *gorm.DB, plan *models.WorkoutPlan) error {
    // Validate plan name is unique
    var existing models.WorkoutPlan
    if err := db.Where("name = ?", plan.Name).First(&existing).Error; err == nil {
        return fmt.Errorf("plan with name '%s' already exists", plan.Name)
    }
    return nil
}
generics.RegisterBasicCRUD(group, db, config)
```

## File Structure

```
backend/
├── generics/
│   ├── generic_service.go    # Existing - data layer CRUD
│   └── handlers.go            # NEW - HTTP handlers
├── workout/
│   └── routes/
│       ├── workout_plan_api.go    # Simplified to ~10 lines
│       ├── exercises_api.go        # Keep custom routes, add CRUD
│       └── workout_log_api.go      # Keep custom routes, add CRUD
```

## Migration Strategy

### Phase 1: Create Generic Handlers
1. Create `generics/handlers.go`
2. Implement all 5 CRUD handlers
3. Add comprehensive tests

### Phase 2: Migrate Simple Entities
1. Start with `WorkoutPlan` (simplest case)
2. Replace existing handlers with generic registration
3. Test thoroughly

### Phase 3: Migrate Other Entities
1. Migrate `Exercise` (if it needs CRUD)
2. Add custom routes alongside CRUD where needed
3. Keep entity-specific handlers for complex operations

### Phase 4: Refinement
1. Add hooks for common patterns (validation, logging)
2. Add middleware support if needed
3. Document best practices

## Benefits

1. **Reduced Code**: 150+ lines → 5-10 lines per entity
2. **Consistency**: All entities behave the same way
3. **Maintainability**: Fix bugs once, applies everywhere
4. **Type Safety**: Generic constraints ensure compile-time safety
5. **Flexibility**: Hooks and config allow customization when needed

## Edge Cases to Handle

1. **Pagination**: Default to paginated (page size 20) if no params, but allow disabling
2. **Preloads**: Use default preloads if entity implements Preloadable
3. **Soft Deletes**: Always use soft delete (Delete, not DeleteHard)
4. **ID Validation**: Validate ID format and existence
5. **JSON Binding**: Handle binding errors gracefully
6. **Custom Filters**: Support query params as filters (e.g., `?name=Push`)

## Testing Strategy

1. **Unit Tests**: Test each handler method with mock DB
2. **Integration Tests**: Test full CRUD flow with real DB
3. **Edge Cases**: Test invalid IDs, missing data, validation errors
4. **Configuration**: Test different config options

## Future Enhancements

1. **Bulk Operations**: Add bulk create/update/delete
2. **Search**: Add full-text search support
3. **Export**: Add CSV/JSON export endpoints
4. **Validation**: Integrate with validation library
5. **Audit Logging**: Add automatic audit trail hooks
