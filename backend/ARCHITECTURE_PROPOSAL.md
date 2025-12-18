# Backend Architecture Proposal

## Current State Analysis

### Current Structure
```
backend/
├── api/
│   ├── api.go              (router init + feature setup)
│   ├── base_api.go         (base feature struct)
│   ├── diet_api.go         (~458 lines - routes + handlers)
│   └── workout_api.go      (~457 lines - routes + handlers)
├── database/
│   ├── database.go         (connection + dump/restore routes)
│   ├── models/
│   │   ├── base_model.go   (interfaces)
│   │   ├── diet_models.go  (entities + migration + seeding)
│   │   └── workout_models.go
│   └── services/
│       ├── diet_services.go     (data access functions)
│       ├── workout_services.go
│       └── generic_service.go
├── utils/
│   └── time.go
└── main.go
```

### Current Issues

| Issue | Description |
|-------|-------------|
| **Monolithic API files** | 450+ line files mixing routes, handlers, request types, and business logic |
| **Models do too much** | Entities contain migration logic, seeding, and preload definitions |
| **Services are functions** | No interfaces, hard to test, no dependency injection |
| **Tight coupling** | `*gorm.DB` passed everywhere, direct database access in handlers |
| **No business layer** | Handlers call services directly, mixing HTTP concerns with domain logic |
| **Scattered DTOs** | Request/Response types defined inline in handler files |

---

## Proposed Architecture: Feature-Based Domain Structure

This proposal uses a **feature-based (vertical slice)** architecture with clean separation of concerns. Each feature is self-contained but shares common infrastructure.

### New Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
│
├── internal/
│   ├── app/
│   │   └── app.go                  # Application bootstrapping, DI container
│   │
│   ├── common/
│   │   ├── response/
│   │   │   └── response.go         # Standard API response helpers
│   │   ├── errors/
│   │   │   └── errors.go           # Application error types
│   │   └── middleware/
│   │       ├── cors.go
│   │       └── logging.go
│   │
│   ├── database/
│   │   ├── database.go             # Connection setup
│   │   ├── migrator.go             # Central migration runner
│   │   └── seeder.go               # Central seeder (dev only)
│   │
│   ├── workout/                    # ──── WORKOUT FEATURE ────
│   │   ├── entity.go               # Pure domain entities (structs only)
│   │   ├── repository.go           # Interface + GORM implementation
│   │   ├── service.go              # Business logic
│   │   ├── handler.go              # HTTP handlers
│   │   ├── routes.go               # Route definitions
│   │   └── dto.go                  # Request/Response types
│   │
│   ├── diet/                       # ──── DIET FEATURE ────
│   │   ├── entity.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── handler.go
│   │   ├── routes.go
│   │   └── dto.go
│   │
│   └── exercise/                   # ──── SHARED SUB-FEATURE ────
│       ├── entity.go               # Exercise is used by workout feature
│       ├── repository.go
│       └── service.go
│
├── pkg/                            # Exportable utilities
│   └── timeutil/
│       └── time.go
│
├── migrations/                     # SQL migrations (optional, for production)
│   └── ...
│
└── go.mod
```

---

## Layer Responsibilities

### 1. Entity Layer (`entity.go`)
Pure Go structs representing domain objects. **No GORM tags for business logic** - only for persistence mapping.

```go
// internal/workout/entity.go
package workout

import "time"

type WorkoutLog struct {
    ID          uint          `gorm:"primaryKey"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Date        time.Time     `json:"date"`
    PlanID      *uint         `json:"workout_plan_id"`
    Plan        *WorkoutPlan  `json:"workout_plan"`
    Exercises   []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
    Cardio      *Cardio       `json:"cardio" gorm:"constraint:OnDelete:CASCADE;"`
}

type WorkoutPlan struct {
    ID        uint       `gorm:"primaryKey"`
    Name      string     `json:"name"`
    Exercises []Exercise `gorm:"many2many:workout_plan_exercises;" json:"exercises"`
}

// ... other entities
```

### 2. Repository Layer (`repository.go`)
Interface defining data access + concrete GORM implementation.

```go
// internal/workout/repository.go
package workout

import (
    "context"
    "time"
    "gorm.io/gorm"
)

// Repository interface - enables testing with mocks
type Repository interface {
    GetByDate(ctx context.Context, date time.Time) (*WorkoutLog, error)
    GetByDateRange(ctx context.Context, start, end time.Time) ([]WorkoutLog, error)
    GetAll(ctx context.Context) ([]WorkoutLog, error)
    Create(ctx context.Context, log *WorkoutLog) error
    Update(ctx context.Context, log *WorkoutLog) error
}

// gormRepository is the GORM implementation
type gormRepository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &gormRepository{db: db}
}

func (r *gormRepository) GetByDate(ctx context.Context, date time.Time) (*WorkoutLog, error) {
    var log WorkoutLog
    err := r.db.WithContext(ctx).
        Preload("Cardio").
        Preload("Exercises.Sets").
        Preload("Exercises.Exercise").
        Preload("Plan.Exercises").
        Where("date = ?", date).
        First(&log).Error
    if err != nil {
        return nil, err
    }
    return &log, nil
}
// ... other methods
```

### 3. Service Layer (`service.go`)
Business logic that orchestrates repositories and applies domain rules.

```go
// internal/workout/service.go
package workout

import (
    "context"
    "be-simpletracker/pkg/timeutil"
)

type Service interface {
    GetTodayWorkout(ctx context.Context, dayOffset int) (*WorkoutLog, error)
    GetMonthWorkouts(ctx context.Context, monthOffset int) (*MonthData, error)
    LogExercise(ctx context.Context, req LogExerciseRequest) error
    // ... other business operations
}

type service struct {
    repo        Repository
    exerciseRepo exercise.Repository  // Cross-feature dependency
}

func NewService(repo Repository, exerciseRepo exercise.Repository) Service {
    return &service{
        repo:         repo,
        exerciseRepo: exerciseRepo,
    }
}

func (s *service) GetTodayWorkout(ctx context.Context, dayOffset int) (*WorkoutLog, error) {
    date := timeutil.ZerodTime(dayOffset)
    return s.repo.GetByDate(ctx, date)
}

// Business logic example: calculate progression
func (s *service) GetExerciseProgression(ctx context.Context, exerciseID uint) (*ProgressionData, error) {
    logs, err := s.exerciseRepo.GetHistoricalLogs(ctx, exerciseID)
    if err != nil {
        return nil, err
    }
    
    // Business logic for calculating volume trends
    return calculateProgression(logs), nil
}
```

### 4. Handler Layer (`handler.go`)
HTTP-specific logic. Parses requests, calls services, formats responses.

```go
// internal/workout/handler.go
package workout

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "be-simpletracker/internal/common/response"
)

type Handler struct {
    svc Service
}

func NewHandler(svc Service) *Handler {
    return &Handler{svc: svc}
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
package workout

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
    g := r.Group("/workout")
    {
        g.GET("/today", h.GetToday)
        g.GET("/month", h.GetMonth)
        g.GET("/all", h.GetAll)
        g.GET("/previous", h.GetPrevious)
        
        // Exercise sub-routes
        ex := g.Group("/exercise")
        {
            ex.POST("/log", h.LogExercise)
            ex.POST("/all-logged", h.CheckAllLogged)
            ex.POST("/add", h.AddExerciseToWorkout)
            ex.DELETE("/remove", h.RemoveExerciseFromWorkout)
            ex.GET("/progression/:id", h.GetExerciseProgression)
        }
        
        // Plan sub-routes  
        plans := g.Group("/plans")
        {
            plans.GET("/all", h.GetAllPlans)
            plans.POST("/:id/exercises/add", h.AddExerciseToPlan)
            plans.DELETE("/:id/exercises/remove", h.RemoveExerciseFromPlan)
        }
    }
}
```

### 6. DTO Layer (`dto.go`)
All request/response types in one place per feature.

```go
// internal/workout/dto.go
package workout

type LogExerciseRequest struct {
    Exercise LoggedExercise `json:"exercise"`
    Type     string         `json:"type"` // "previous" | "logged"
}

type AddExerciseRequest struct {
    ExerciseID uint `json:"exercise_id" binding:"required"`
}

type MonthResponse struct {
    Days   []WorkoutLog `json:"days"`
    Today  time.Time    `json:"today"`
    Range  DateRange    `json:"range"`
    Month  time.Month   `json:"month"`
    Offset int          `json:"offset"`
}

type DateRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}
```

---

## Application Bootstrap (`internal/app/app.go`)

Central dependency injection and feature registration:

```go
package app

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "be-simpletracker/internal/database"
    "be-simpletracker/internal/workout"
    "be-simpletracker/internal/diet"
    "be-simpletracker/internal/exercise"
    "be-simpletracker/internal/common/middleware"
)

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
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    
    app := &App{Router: router, DB: db}
    app.registerFeatures()
    
    return app, nil
}

func (a *App) registerFeatures() {
    // Shared repositories
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

func (a *App) Run(addr string) error {
    return a.Router.Run(addr)
}
```

---

## Breaking Down Large Features (Nested Sub-Domains)

Your current workout feature handles:
- Workout day/log management
- Exercise logging  
- Workout plans
- Exercise progression

When a feature grows complex, you can nest **sub-domains** within it. Each sub-domain owns its own handler, service, repository, and entities.

### Nested Structure Example

```
internal/
├── workout/
│   ├── workout.go              # Feature-level bootstrap (wires sub-domains)
│   ├── shared/                 # Shared entities used across sub-domains
│   │   └── entity.go           # WorkoutLog, LoggedExercise, etc.
│   │
│   ├── log/                    # ──── Daily workout logging ────
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── routes.go
│   │   └── dto.go
│   │
│   ├── plan/                   # ──── Workout plan management ────
│   │   ├── entity.go           # WorkoutPlan (owned by this sub-domain)
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── routes.go
│   │   └── dto.go
│   │
│   └── progression/            # ──── Historical analysis ────
│       ├── handler.go
│       ├── service.go          # Complex calculation logic
│       ├── routes.go
│       └── dto.go              # ProgressionEntry, ChartData, etc.
│
├── exercise/                   # Standalone feature (shared across workout)
│   ├── entity.go
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   ├── routes.go
│   └── dto.go
│
└── diet/
    ├── diet.go                 # Feature bootstrap
    ├── shared/
    │   └── entity.go           # Day, DayLog (shared)
    │
    ├── day/                    # Daily meal tracking
    │   ├── handler.go
    │   ├── service.go
    │   ├── repository.go
    │   ├── routes.go
    │   └── dto.go
    │
    ├── meal/                   # Meal definitions
    │   ├── entity.go           # Meal, MealItem
    │   ├── handler.go
    │   ├── service.go
    │   ├── repository.go
    │   ├── routes.go
    │   └── dto.go
    │
    ├── food/                   # Food database
    │   ├── entity.go           # Food
    │   ├── handler.go
    │   ├── service.go
    │   ├── repository.go
    │   ├── routes.go
    │   └── dto.go
    │
    └── goal/                   # Nutrition goals/plans
        ├── entity.go           # Plan (calorie/protein targets)
        ├── handler.go
        ├── service.go
        ├── repository.go
        ├── routes.go
        └── dto.go
```

---

## When to Nest vs. Keep Flat

| Nest (sub-domains) | Keep Flat |
|--------------------|-----------|
| Sub-feature has 5+ endpoints | Simple CRUD with 2-3 endpoints |
| Complex business logic specific to that area | Logic is straightforward |
| Could potentially become its own microservice | Tightly coupled to parent feature |
| Different team members work on different parts | Single developer maintains it |
| Entity is "owned" by the sub-domain | Entity is shared across feature |

### Your Workout Feature Analysis

| Sub-Domain | Endpoints | Complexity | Recommendation |
|------------|-----------|------------|----------------|
| **log** | 3 (today, month, all) | Medium - date ranges, preloads | ✅ Nest |
| **plan** | 4+ (all, add exercise, remove) | Medium - many2many relations | ✅ Nest |
| **exercise** | 5+ (log, add, remove, progression) | High - cross-references logs | ✅ Separate feature |
| **progression** | 1-2 | High - calculations | 🤔 Maybe later |

---

## Sub-Domain Wiring (`workout/workout.go`)

The parent feature wires up all sub-domains:

```go
// internal/workout/workout.go
package workout

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "be-simpletracker/internal/workout/log"
    "be-simpletracker/internal/workout/plan"
    "be-simpletracker/internal/exercise"
)

type Feature struct {
    logHandler  *log.Handler
    planHandler *plan.Handler
}

func NewFeature(db *gorm.DB, exerciseRepo exercise.Repository) *Feature {
    // Log sub-domain
    logRepo := log.NewRepository(db)
    logSvc := log.NewService(logRepo, exerciseRepo)
    logHandler := log.NewHandler(logSvc)
    
    // Plan sub-domain
    planRepo := plan.NewRepository(db)
    planSvc := plan.NewService(planRepo, exerciseRepo)
    planHandler := plan.NewHandler(planSvc)
    
    return &Feature{
        logHandler:  logHandler,
        planHandler: planHandler,
    }
}

func (f *Feature) RegisterRoutes(r *gin.Engine) {
    g := r.Group("/workout")
    
    log.RegisterRoutes(g, f.logHandler)     // /workout/today, /workout/month
    plan.RegisterRoutes(g, f.planHandler)   // /workout/plans/...
}
```

---

## Sub-Domain Routes Example

```go
// internal/workout/plan/routes.go
package plan

import "github.com/gin-gonic/gin"

func RegisterRoutes(parent *gin.RouterGroup, h *Handler) {
    g := parent.Group("/plans")
    {
        g.GET("/all", h.GetAll)
        g.GET("/:id", h.GetByID)
        g.POST("/", h.Create)
        g.PUT("/:id", h.Update)
        g.DELETE("/:id", h.Delete)
        
        // Exercises within a plan
        g.POST("/:id/exercises", h.AddExercise)
        g.DELETE("/:id/exercises/:exerciseId", h.RemoveExercise)
    }
}
```

```go
// internal/workout/log/routes.go
package log

import "github.com/gin-gonic/gin"

func RegisterRoutes(parent *gin.RouterGroup, h *Handler) {
    // These attach directly to /workout
    parent.GET("/today", h.GetToday)
    parent.GET("/month", h.GetMonth)
    parent.GET("/all", h.GetAll)
    parent.GET("/previous", h.GetPrevious)
    
    // Exercise logging within daily workout
    ex := parent.Group("/exercise")
    {
        ex.POST("/log", h.LogExercise)
        ex.POST("/add", h.AddExercise)
        ex.DELETE("/remove", h.RemoveExercise)
    }
}
```

---

## Shared Entities Strategy

When entities are used across sub-domains, put them in a `shared/` folder:

```go
// internal/workout/shared/entity.go
package shared

import "time"

type WorkoutLog struct {
    ID          uint             `gorm:"primaryKey"`
    Date        time.Time        `json:"date"`
    PlanID      *uint            `json:"workout_plan_id"`
    Exercises   []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
}

type LoggedExercise struct {
    ID           uint        `gorm:"primaryKey"`
    WorkoutLogID uint        `json:"workout_log_id"`
    ExerciseID   uint        `json:"exercise_id"`
    Sets         []LoggedSet `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
}

type LoggedSet struct {
    ID               uint    `gorm:"primaryKey"`
    LoggedExerciseID uint    `json:"logged_exercise_id"`
    Reps             uint    `json:"reps"`
    Weight           float32 `json:"weight"`
}
```

Sub-domains import from shared:

```go
// internal/workout/log/repository.go
package log

import (
    "be-simpletracker/internal/workout/shared"
    "gorm.io/gorm"
)

type Repository interface {
    GetByDate(date time.Time) (*shared.WorkoutLog, error)
}
```

---

## Entity Ownership Rule

**Each entity should have ONE owner** (the sub-domain that creates/manages it):

| Entity | Owner | Used By |
|--------|-------|---------|
| `WorkoutLog` | `workout/log` | `workout/progression` |
| `WorkoutPlan` | `workout/plan` | `workout/log` (references) |
| `Exercise` | `exercise` (standalone) | Everyone |
| `LoggedExercise` | `workout/log` | `workout/progression` |
| `Day` | `diet/day` | `diet/meal` |
| `Meal` | `diet/meal` | `diet/day` |
| `Food` | `diet/food` | `diet/meal` |
| `Plan` (nutrition) | `diet/goal` | `diet/day` |

The owner is responsible for:
- The entity definition
- Migrations for that entity
- Basic CRUD repository methods

Other sub-domains can **read** but should call the owner's service for **writes**.

---

## Migration Strategy

### Phase 1: Extract DTOs (Low Risk)
Move all request/response types to `dto.go` files within each feature.

### Phase 2: Introduce Repositories (Medium Risk)
Create repository interfaces alongside existing services. Migrate service functions to use them.

### Phase 3: Split Handler Files (Low Risk)
Move route definitions to `routes.go`, keep handlers in `handler.go`.

### Phase 4: Restructure Folders (Medium Risk)
Move to the new `internal/` structure. Update imports.

### Phase 5: Clean Up Models (Medium Risk)
- Move entities to feature `entity.go` files
- Extract migration/seeding to `internal/database/`

---

## Benefits of This Architecture

| Benefit | Description |
|---------|-------------|
| **Testability** | Repository interfaces enable mocking for unit tests |
| **Scalability** | Add new features without touching existing code |
| **Maintainability** | Small, focused files (~100-200 lines each) |
| **Clarity** | Clear flow: Route → Handler → Service → Repository |
| **Flexibility** | Swap database implementations without changing business logic |
| **Team Scaling** | Multiple developers can work on different features |

---

## Common Response Helpers

```go
// internal/common/response/response.go
package response

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

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

## Quick Wins (Do First)

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
