package generics

import (
	"be-simpletracker/database/repository"
	"context"
	"time"

	"gorm.io/gorm"
)

// ========== Generic CRUD Operations ==========
// These functions provide a simple, consistent interface for all entities
// that implement the repository.Entity interface

// GetAll retrieves all entities matching the query options
// Usage: entities, err := generics.GetAll[models.WorkoutPlan](ctx, db, repository.WithDefaultPreloads())
func GetAll[T repository.Entity](ctx context.Context, db *gorm.DB, opts ...repository.QueryOption) ([]T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetAll(ctx, opts...)
}

// GetOne retrieves a single entity by ID
// This is an alias for GetByID for consistency with common naming conventions
// Usage: entity, err := generics.GetOne[models.WorkoutPlan](ctx, db, id, repository.WithDefaultPreloads())
func GetOne[T repository.Entity](ctx context.Context, db *gorm.DB, id uint, opts ...repository.QueryOption) (T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetByID(ctx, id, opts...)
}

// GetByID retrieves a single entity by ID
// Usage: entity, err := generics.GetByID[models.WorkoutPlan](ctx, db, id, repository.WithDefaultPreloads())
func GetByID[T repository.Entity](ctx context.Context, db *gorm.DB, id uint, opts ...repository.QueryOption) (T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetByID(ctx, id, opts...)
}

// FindOne finds a single entity matching the query options (filters)
// Usage: entity, err := generics.FindOne[models.WorkoutPlan](ctx, db, repository.WithFilter("name", "Push"))
func FindOne[T repository.Entity](ctx context.Context, db *gorm.DB, opts ...repository.QueryOption) (T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.FindOne(ctx, opts...)
}

// Create creates a new entity
// Usage: err := generics.Create(ctx, db, entity)
func Create[T repository.Entity](ctx context.Context, db *gorm.DB, entity T) error {
	repo := repository.NewGormRepository[T](db)
	return repo.Create(ctx, entity)
}

// Update updates an existing entity
// Usage: err := generics.Update(ctx, db, entity)
func Update[T repository.Entity](ctx context.Context, db *gorm.DB, entity T) error {
	repo := repository.NewGormRepository[T](db)
	return repo.Update(ctx, entity)
}

// Delete deletes an entity by ID (soft delete if supported)
// Usage: err := generics.Delete[models.WorkoutPlan](ctx, db, id)
func Delete[T repository.Entity](ctx context.Context, db *gorm.DB, id uint) error {
	repo := repository.NewGormRepository[T](db)
	return repo.Delete(ctx, id)
}

// ========== Additional Helper Functions ==========

// CreateEntity is an alias for Create (for backward compatibility)
func CreateEntity[T repository.Entity](ctx context.Context, db *gorm.DB, entity T) error {
	return Create(ctx, db, entity)
}

// UpdateEntity is an alias for Update (for backward compatibility)
func UpdateEntity[T repository.Entity](ctx context.Context, db *gorm.DB, entity T) error {
	return Update(ctx, db, entity)
}

// DeleteEntity is an alias for Delete (for backward compatibility)
func DeleteEntity[T repository.Entity](ctx context.Context, db *gorm.DB, id uint) error {
	return Delete[T](ctx, db, id)
}

// DeleteHard permanently deletes an entity by ID
// Usage: err := generics.DeleteHard[models.WorkoutPlan](ctx, db, id)
func DeleteHard[T repository.Entity](ctx context.Context, db *gorm.DB, id uint) error {
	repo := repository.NewGormRepository[T](db)
	return repo.DeleteHard(ctx, id)
}

// Exists checks if an entity with the given ID exists
// Usage: exists, err := generics.Exists[models.WorkoutPlan](ctx, db, id)
func Exists[T repository.Entity](ctx context.Context, db *gorm.DB, id uint) (bool, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.Exists(ctx, id)
}

// ExistsEntity is an alias for Exists (for backward compatibility)
func ExistsEntity[T repository.Entity](ctx context.Context, db *gorm.DB, id uint) (bool, error) {
	return Exists[T](ctx, db, id)
}

// GetAllPaginated retrieves entities with pagination metadata
// Usage: result, err := generics.GetAllPaginated[models.WorkoutPlan](ctx, db, 1, 10, repository.WithDefaultPreloads())
func GetAllPaginated[T repository.Entity](ctx context.Context, db *gorm.DB, page, pageSize int, opts ...repository.QueryOption) (*repository.PaginatedResult[T], error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetAllPaginated(ctx, page, pageSize, opts...)
}

// GetByDateRange retrieves entities within a date range (for Dateable entities)
// Usage: entities, err := generics.GetByDateRange[models.WorkoutLog](ctx, db, start, end, repository.WithDefaultPreloads())
func GetByDateRange[T repository.Entity](ctx context.Context, db *gorm.DB, start, end time.Time, opts ...repository.QueryOption) ([]T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetByDateRange(ctx, start, end, opts...)
}

// Count returns the total count of entities matching filters
// Usage: count, err := generics.Count[models.WorkoutPlan](ctx, db, repository.WithFilter("name", "Push"))
func Count[T repository.Entity](ctx context.Context, db *gorm.DB, opts ...repository.QueryOption) (int64, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.Count(ctx, opts...)
}

// NewRepository creates a generic repository for any entity type
// This is a convenience function for creating repositories
// Usage: repo := generics.NewRepository[models.WorkoutPlan](db)
func NewRepository[T repository.Entity](db *gorm.DB) *repository.GormRepository[T] {
	return repository.NewGormRepository[T](db)
}

// NewDateableRepository creates a repository with date field support
// Usage: repo := generics.NewDateableRepository[models.WorkoutLog](db, "date")
func NewDateableRepository[T repository.Entity](db *gorm.DB, dateField string) *repository.GormRepository[T] {
	return repository.NewGormRepositoryWithDateField[T](db, dateField)
}