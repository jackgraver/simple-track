package repository

import (
	"context"
	"time"
)

// Repository defines the generic repository interface for CRUD operations
// T must be a pointer to a struct that implements Entity
type Repository[T Entity] interface {
	// GetByID retrieves a single entity by ID
	GetByID(ctx context.Context, id uint, opts ...QueryOption) (T, error)

	// GetAll retrieves all entities matching the query options
	GetAll(ctx context.Context, opts ...QueryOption) ([]T, error)

	// GetAllPaginated retrieves entities with pagination metadata
	GetAllPaginated(ctx context.Context, page, pageSize int, opts ...QueryOption) (*PaginatedResult[T], error)

	// Count returns the total count of entities matching filters
	Count(ctx context.Context, opts ...QueryOption) (int64, error)

	// Create inserts a new entity
	Create(ctx context.Context, entity T) error

	// Update updates an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity by ID (soft delete if supported)
	Delete(ctx context.Context, id uint) error

	// DeleteHard permanently removes an entity by ID
	DeleteHard(ctx context.Context, id uint) error

	// Exists checks if an entity with the given ID exists
	Exists(ctx context.Context, id uint) (bool, error)

	// FindOne finds a single entity matching filters
	FindOne(ctx context.Context, opts ...QueryOption) (T, error)

	// Transaction executes a function within a transaction
	Transaction(ctx context.Context, fn func(repo Repository[T]) error) error
}

// DateableRepository extends Repository with date-based query methods
type DateableRepository[T Entity] interface {
	Repository[T]

	// GetByDate retrieves entities for a specific date
	GetByDate(ctx context.Context, date time.Time, opts ...QueryOption) (T, error)

	// GetByDateRange retrieves entities within a date range
	GetByDateRange(ctx context.Context, start, end time.Time, opts ...QueryOption) ([]T, error)

	// GetByDateRangePaginated retrieves entities within a date range with pagination
	GetByDateRangePaginated(ctx context.Context, start, end time.Time, page, pageSize int, opts ...QueryOption) (*PaginatedResult[T], error)
}

// BatchRepository adds batch operation support
type BatchRepository[T Entity] interface {
	Repository[T]

	// CreateBatch inserts multiple entities
	CreateBatch(ctx context.Context, entities []T) error

	// UpdateBatch updates multiple entities
	UpdateBatch(ctx context.Context, entities []T) error

	// DeleteBatch deletes multiple entities by IDs
	DeleteBatch(ctx context.Context, ids []uint) error
}
