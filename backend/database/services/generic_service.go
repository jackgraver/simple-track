package services

import (
	"be-simpletracker/database/models"
	"be-simpletracker/database/repository"
	"context"
	"time"

	"gorm.io/gorm"
)

// ObjectRange retrieves entities within a date range using the new repository pattern
// Deprecated: Use repository.GormRepository.GetByDateRange instead
func ObjectRange[T models.Preloadable](db *gorm.DB, start time.Time, end time.Time) ([]T, error) {
	var objects []T
	var t T

	tx := db
	for _, p := range t.Preloads() {
		tx = tx.Preload(p)
	}

	if err := tx.
		Where("date BETWEEN ? AND ?", start, end).
		Order("date").
		Find(&objects).Error; err != nil {
		return nil, err
	}
	return objects, nil
}

// NewRepository creates a generic repository for any entity type
// This is a convenience function for creating repositories
func NewRepository[T repository.Entity](db *gorm.DB) *repository.GormRepository[T] {
	return repository.NewGormRepository[T](db)
}

// NewDateableRepository creates a repository with date field support
func NewDateableRepository[T repository.Entity](db *gorm.DB, dateField string) *repository.GormRepository[T] {
	return repository.NewGormRepositoryWithDateField[T](db, dateField)
}

// GetByDateRange is a convenience function using the new repository
func GetByDateRange[T repository.Entity](db *gorm.DB, start, end time.Time, opts ...repository.QueryOption) ([]T, error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetByDateRange(context.Background(), start, end, opts...)
}

// GetAllPaginated is a convenience function for paginated queries
func GetAllPaginated[T repository.Entity](db *gorm.DB, page, pageSize int, opts ...repository.QueryOption) (*repository.PaginatedResult[T], error) {
	repo := repository.NewGormRepository[T](db)
	return repo.GetAllPaginated(context.Background(), page, pageSize, opts...)
}