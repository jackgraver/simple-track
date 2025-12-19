# Backend Architecture Proposal

## Current State Analysis

### Current Structure
```
backend/
├── api/ (diet_api.go, workout_api.go - 450+ lines each)
├── database/
│   ├── models/ (entities + migration + seeding)
│   └── services/ (data access functions)
└── utils/
```

### Current Issues
| Issue | Description |
|-------|-------------|
| **Monolithic API files** | 450+ line files mixing routes, handlers, DTOs, and business logic |
| **Models do too much** | Entities contain migration logic, seeding, and preload definitions |
| **Services are functions** | No interfaces, hard to test, no dependency injection |
| **Tight coupling** | `*gorm.DB` passed everywhere, direct database access in handlers |
| **No business layer** | Handlers call services directly, mixing HTTP concerns with domain logic |
| **Scattered DTOs** | Request/Response types defined inline in handler files |

---

## Proposed Architecture: Feature-Based Domain Structure

Feature-based (vertical slice) architecture with clean separation of concerns. Each feature is self-contained but shares common infrastructure.

### New Structure
```
backend/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── app/app.go                  # Application bootstrapping, DI container
│   ├── common/
│   │   ├── response/response.go    # Standard API response helpers
│   │   ├── errors/errors.go        # Application error types
│   │   └── middleware/             # CORS, logging
│   ├── database/
│   │   ├── database.go             # Connection setup
│   │   ├── migrator.go             # Central migration runner
│   │   └── seeder.go               # Central seeder (dev only)
│   ├── workout/                    # ──── WORKOUT FEATURE ────
│   │   ├── entity.go               # Pure domain entities
│   │   ├── repository.go           # Interface + GORM implementation
│   │   ├── service.go              # Business logic
│   │   ├── handler.go              # HTTP handlers
│   │   ├── routes.go               # Route definitions
│   │   └── dto.go                  # Request/Response types
│   ├── diet/                       # ──── DIET FEATURE ────
│   │   └── [same structure]
│   └── exercise/                   # ──── SHARED FEATURE ────
│       └── [entity, repository, service]
├── pkg/timeutil/                   # Exportable utilities
└── migrations/                     # SQL migrations (optional)
```

---

## Layer Responsibilities

### 1. Entity Layer (`entity.go`)
Pure Go structs representing domain objects. Only GORM tags for persistence mapping.

```go
// internal/workout/entity.go
type WorkoutLog struct {
    ID        uint          `gorm:"primaryKey"`
    Date      time.Time     `json:"date"`
    PlanID    *uint         `json:"workout_plan_id"`
    Plan      *WorkoutPlan  `json:"workout_plan"`
    Exercises []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
}
```

### 2. Repository Layer (`repository.go`)
Interface defining data access + concrete GORM implementation. Enables testing with mocks.

```go
// internal/workout/repository.go
type Repository interface {
    GetByDate(ctx context.Context, date time.Time) (*WorkoutLog, error)
    GetByDateRange(ctx context.Context, start, end time.Time) ([]WorkoutLog, error)
    Create(ctx context.Context, log *WorkoutLog) error
    Update(ctx context.Context, log *WorkoutLog) error
}

type gormRepository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &gormRepository{db: db}
}
```

### 3. Service Layer (`service.go`)
Business logic that orchestrates repositories and applies domain rules.

```go
// internal/workout/service.go
type Service interface {
    GetTodayWorkout(ctx context.Context, dayOffset int) (*WorkoutLog, error)
    LogExercise(ctx context.Context, req LogExerciseRequest) error
}

type service struct {
    repo        Repository
    exerciseRepo exercise.Repository  // Cross-feature dependency
}

func NewService(repo Repository, exerciseRepo exercise.Repository) Service {
    return &service{repo: repo, exerciseRepo: exerciseRepo}
}
```

### 4. Handler Layer (`handler.go`)
HTTP-specific logic. Parses requests, calls services, formats responses.

```go
// internal/workout/handler.go
type Handler struct {
    svc Service
}

func (h *Handler) GetToday(c *gin.Context) {
    offset, _ := strconv.Atoi(c.Query("offset"))
    day, err := h.svc.GetTodayWorkout(c.Request.Context(), offset)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err)
        return
    }
    response.Success(c, day)
}
```

### 5. Routes Layer (`routes.go`)
Clean route registration, separate from handler logic.

```go
// internal/workout/routes.go
func RegisterRoutes(r *gin.Engine, h *Handler) {
    g := r.Group("/workout")
    {
        g.GET("/today", h.GetToday)
        g.GET("/month", h.GetMonth)
        ex := g.Group("/exercise")
        {
            ex.POST("/log", h.LogExercise)
            ex.DELETE("/remove", h.RemoveExercise)
        }
    }
}
```

### 6. DTO Layer (`dto.go`)
All request/response types in one place per feature.

```go
// internal/workout/dto.go
type LogExerciseRequest struct {
    Exercise LoggedExercise `json:"exercise"`
    Type     string         `json:"type"` // "previous" | "logged"
}

type MonthResponse struct {
    Days  []WorkoutLog `json:"days"`
    Today time.Time    `json:"today"`
}
```

---

## Application Bootstrap (`internal/app/app.go`)

Central dependency injection and feature registration:

```go
// internal/app/app.go
type App struct {
    Router *gin.Engine
    DB     *gorm.DB
}

func New() (*App, error) {
    db, err := database.Connect()
    if err != nil {
        return nil, err
    }
    
    router := gin.Default()
    router.Use(middleware.CORS(), middleware.Logger())
    
    app := &App{Router: router, DB: db}
    app.registerFeatures()
    return app, nil
}

func (a *App) registerFeatures() {
    exerciseRepo := exercise.NewRepository(a.DB)
    
    // Workout feature
    workoutRepo := workout.NewRepository(a.DB)
    workoutSvc := workout.NewService(workoutRepo, exerciseRepo)
    workoutHandler := workout.NewHandler(workoutSvc)
    workout.RegisterRoutes(a.Router, workoutHandler)
    
    // Diet feature
    dietRepo := diet.NewRepository(a.DB)
    dietSvc := diet.NewService(dietRepo)
    dietHandler := diet.NewHandler(dietSvc)
    diet.RegisterRoutes(a.Router, dietHandler)
}
```

---

## Breaking Down Large Features (Nested Sub-Domains)

When a feature grows complex, nest **sub-domains** within it. Each sub-domain owns its own handler, service, repository, and entities.

### Nested Structure Example
```
internal/
├── workout/
│   ├── workout.go              # Feature-level bootstrap (wires sub-domains)
│   ├── shared/entity.go        # WorkoutLog, LoggedExercise (shared)
│   ├── log/                    # Daily workout logging
│   │   └── [handler, service, repository, routes, dto]
│   ├── plan/                   # Workout plan management
│   │   └── [entity, handler, service, repository, routes, dto]
│   └── progression/            # Historical analysis
│       └── [handler, service, routes, dto]
└── diet/
    ├── shared/entity.go        # Day, DayLog (shared)
    ├── day/                    # Daily meal tracking
    ├── meal/                   # Meal definitions
    ├── food/                   # Food database
    └── goal/                   # Nutrition goals/plans
```

### When to Nest vs. Keep Flat

| Nest (sub-domains) | Keep Flat |
|--------------------|-----------|
| Sub-feature has 5+ endpoints | Simple CRUD with 2-3 endpoints |
| Complex business logic specific to that area | Logic is straightforward |
| Could potentially become its own microservice | Tightly coupled to parent feature |
| Different team members work on different parts | Single developer maintains it |
| Entity is "owned" by the sub-domain | Entity is shared across feature |

**Workout Feature Analysis:**
- **log**: 3 endpoints, medium complexity → ✅ Nest
- **plan**: 4+ endpoints, medium complexity → ✅ Nest
- **exercise**: 5+ endpoints, high complexity → ✅ Separate feature
- **progression**: 1-2 endpoints, high complexity → 🤔 Maybe later

### Sub-Domain Wiring Example

```go
// internal/workout/workout.go
func NewFeature(db *gorm.DB, exerciseRepo exercise.Repository) *Feature {
    logRepo := log.NewRepository(db)
    logSvc := log.NewService(logRepo, exerciseRepo)
    logHandler := log.NewHandler(logSvc)
    
    planRepo := plan.NewRepository(db)
    planSvc := plan.NewService(planRepo, exerciseRepo)
    planHandler := plan.NewHandler(planSvc)
    
    return &Feature{logHandler: logHandler, planHandler: planHandler}
}

func (f *Feature) RegisterRoutes(r *gin.Engine) {
    g := r.Group("/workout")
    log.RegisterRoutes(g, f.logHandler)     // /workout/today, /workout/month
    plan.RegisterRoutes(g, f.planHandler)   // /workout/plans/...
}
```

### Shared Entities Strategy

When entities are used across sub-domains, put them in a `shared/` folder. Sub-domains import from shared:

```go
// internal/workout/shared/entity.go
package shared

type WorkoutLog struct {
    ID        uint             `gorm:"primaryKey"`
    Date      time.Time        `json:"date"`
    Exercises []LoggedExercise `json:"exercises"`
}

// internal/workout/log/repository.go
import "be-simpletracker/internal/workout/shared"

type Repository interface {
    GetByDate(date time.Time) (*shared.WorkoutLog, error)
}
```

### Entity Ownership Rule

**Each entity should have ONE owner** (the sub-domain that creates/manages it):

| Entity | Owner | Used By |
|--------|-------|---------|
| `WorkoutLog` | `workout/log` | `workout/progression` |
| `WorkoutPlan` | `workout/plan` | `workout/log` (references) |
| `Exercise` | `exercise` (standalone) | Everyone |
| `Day` | `diet/day` | `diet/meal` |
| `Meal` | `diet/meal` | `diet/day` |

The owner is responsible for: entity definition, migrations, and basic CRUD repository methods. Other sub-domains can **read** but should call the owner's service for **writes**.

---

## Migration Strategy

1. **Phase 1: Extract DTOs** (Low Risk) - Move all request/response types to `dto.go` files
2. **Phase 2: Introduce Repositories** (Medium Risk) - Create repository interfaces alongside existing services
3. **Phase 3: Split Handler Files** (Low Risk) - Move route definitions to `routes.go`
4. **Phase 4: Restructure Folders** (Medium Risk) - Move to the new `internal/` structure
5. **Phase 5: Clean Up Models** (Medium Risk) - Move entities to feature `entity.go` files, extract migration/seeding

---

## Common Response Helpers

```go
// internal/common/response/response.go
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, data)
}

func Error(c *gin.Context, status int, err error) {
    c.JSON(status, gin.H{"error": err.Error()})
}

func NotFound(c *gin.Context, message string) {
    c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func BadRequest(c *gin.Context, err error) {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
```

---

## Benefits & Quick Wins

### Benefits
| Benefit | Description |
|---------|-------------|
| **Testability** | Repository interfaces enable mocking for unit tests |
| **Scalability** | Add new features without touching existing code |
| **Maintainability** | Small, focused files (~100-200 lines each) |
| **Clarity** | Clear flow: Route → Handler → Service → Repository |
| **Flexibility** | Swap database implementations without changing business logic |
| **Team Scaling** | Multiple developers can work on different features |

### Quick Wins (Do First)
1. **Create `dto.go` files** - Extract all request/response types from API files
2. **Create `routes.go` files** - Move `SetEndpoints` logic to dedicated files
3. **Move seeding to a separate file** - Remove `seedDatabase()` from model files
4. **Standardize response format** - Create common response helpers

These changes require minimal refactoring but dramatically improve readability.

---

## Future Considerations

- **User authentication** - Add `internal/auth/` feature with JWT middleware
- **Caching** - Repository decorator pattern for Redis caching
- **API versioning** - Route groups like `/api/v1/workout/`
- **OpenAPI/Swagger** - Generate from DTOs with annotations
- **Background jobs** - Separate `internal/jobs/` for scheduled tasks
