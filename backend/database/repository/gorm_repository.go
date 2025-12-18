package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// GormRepository is the GORM implementation of the generic Repository interface
type GormRepository[T Entity] struct {
	db        *gorm.DB
	dateField string // Field name for date queries (e.g., "date", "created_at")
}

// NewGormRepository creates a new GORM-backed repository
func NewGormRepository[T Entity](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{
		db:        db,
		dateField: "date", // default date field
	}
}

// NewGormRepositoryWithDateField creates a repository with a custom date field
func NewGormRepositoryWithDateField[T Entity](db *gorm.DB, dateField string) *GormRepository[T] {
	return &GormRepository[T]{
		db:        db,
		dateField: dateField,
	}
}

// DB returns the underlying gorm.DB for advanced queries
func (r *GormRepository[T]) DB() *gorm.DB {
	return r.db
}

// applyOptions applies QueryOptions to a gorm.DB query
func (r *GormRepository[T]) applyOptions(tx *gorm.DB, entity T, opts *QueryOptions) *gorm.DB {
	// Apply preloads
	if opts.UseDefaults {
		if p, ok := any(entity).(Preloadable); ok {
			for _, preload := range p.Preloads() {
				tx = tx.Preload(preload)
			}
		}
	} else if len(opts.Preloads) > 0 {
		for _, preload := range opts.Preloads {
			tx = tx.Preload(preload)
		}
	}

	// Apply filters
	for field, value := range opts.Filters {
		tx = tx.Where(field+" = ?", value)
	}

	// Apply exclude IDs
	if len(opts.ExcludeIDs) > 0 {
		tx = tx.Where("id NOT IN ?", opts.ExcludeIDs)
	}

	// Apply date range
	if opts.DateStart != nil {
		tx = tx.Where(r.dateField+" >= ?", *opts.DateStart)
	}
	if opts.DateEnd != nil {
		tx = tx.Where(r.dateField+" <= ?", *opts.DateEnd)
	}

	// Apply ordering
	if opts.OrderBy != "" {
		order := opts.OrderBy
		if opts.OrderDesc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		tx = tx.Order(order)
	}

	// Apply pagination
	if opts.Limit > 0 {
		tx = tx.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		tx = tx.Offset(opts.Offset)
	}

	return tx
}

// GetByID retrieves a single entity by ID
func (r *GormRepository[T]) GetByID(ctx context.Context, id uint, opts ...QueryOption) (T, error) {
	var entity T
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx)
	tx = r.applyOptions(tx, entity, options)

	if err := tx.First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity, ErrNotFound
		}
		return entity, err
	}
	return entity, nil
}

// GetAll retrieves all entities matching the query options
func (r *GormRepository[T]) GetAll(ctx context.Context, opts ...QueryOption) ([]T, error) {
	var entities []T
	var sample T
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx)
	tx = r.applyOptions(tx, sample, options)

	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// GetAllPaginated retrieves entities with pagination metadata
func (r *GormRepository[T]) GetAllPaginated(ctx context.Context, page, pageSize int, opts ...QueryOption) (*PaginatedResult[T], error) {
	var entities []T
	var sample T
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	options := ApplyOptions(opts...)

	// Count total (without limit/offset)
	countTx := r.db.WithContext(ctx).Model(&sample)
	for field, value := range options.Filters {
		countTx = countTx.Where(field+" = ?", value)
	}
	if len(options.ExcludeIDs) > 0 {
		countTx = countTx.Where("id NOT IN ?", options.ExcludeIDs)
	}
	if options.DateStart != nil {
		countTx = countTx.Where(r.dateField+" >= ?", *options.DateStart)
	}
	if options.DateEnd != nil {
		countTx = countTx.Where(r.dateField+" <= ?", *options.DateEnd)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination to options
	options.Limit = pageSize
	options.Offset = (page - 1) * pageSize

	tx := r.db.WithContext(ctx)
	tx = r.applyOptions(tx, sample, options)

	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	return NewPaginatedResult(entities, total, page, pageSize), nil
}

// Count returns the total count of entities matching filters
func (r *GormRepository[T]) Count(ctx context.Context, opts ...QueryOption) (int64, error) {
	var sample T
	var count int64
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx).Model(&sample)

	for field, value := range options.Filters {
		tx = tx.Where(field+" = ?", value)
	}
	if len(options.ExcludeIDs) > 0 {
		tx = tx.Where("id NOT IN ?", options.ExcludeIDs)
	}
	if options.DateStart != nil {
		tx = tx.Where(r.dateField+" >= ?", *options.DateStart)
	}
	if options.DateEnd != nil {
		tx = tx.Where(r.dateField+" <= ?", *options.DateEnd)
	}

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Create inserts a new entity
func (r *GormRepository[T]) Create(ctx context.Context, entity T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update updates an existing entity
func (r *GormRepository[T]) Update(ctx context.Context, entity T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete removes an entity by ID (soft delete if the entity has DeletedAt)
func (r *GormRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// DeleteHard permanently removes an entity by ID
func (r *GormRepository[T]) DeleteHard(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error
}

// Exists checks if an entity with the given ID exists
func (r *GormRepository[T]) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindOne finds a single entity matching filters
func (r *GormRepository[T]) FindOne(ctx context.Context, opts ...QueryOption) (T, error) {
	var entity T
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx)
	tx = r.applyOptions(tx, entity, options)

	if err := tx.First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity, ErrNotFound
		}
		return entity, err
	}
	return entity, nil
}

// Transaction executes a function within a transaction
func (r *GormRepository[T]) Transaction(ctx context.Context, fn func(repo Repository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &GormRepository[T]{db: tx, dateField: r.dateField}
		return fn(txRepo)
	})
}

// ========== DateableRepository methods ==========

// GetByDate retrieves a single entity for a specific date
func (r *GormRepository[T]) GetByDate(ctx context.Context, date time.Time, opts ...QueryOption) (T, error) {
	var entity T
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx)
	tx = r.applyOptions(tx, entity, options)
	tx = tx.Where(r.dateField+" = ?", date)

	if err := tx.First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity, ErrNotFound
		}
		return entity, err
	}
	return entity, nil
}

// GetByDateRange retrieves entities within a date range
func (r *GormRepository[T]) GetByDateRange(ctx context.Context, start, end time.Time, opts ...QueryOption) ([]T, error) {
	// Add date range to options
	opts = append(opts, WithDateRange(start, end), WithOrderByAsc(r.dateField))
	return r.GetAll(ctx, opts...)
}

// GetByDateRangePaginated retrieves entities within a date range with pagination
func (r *GormRepository[T]) GetByDateRangePaginated(ctx context.Context, start, end time.Time, page, pageSize int, opts ...QueryOption) (*PaginatedResult[T], error) {
	opts = append(opts, WithDateRange(start, end), WithOrderByAsc(r.dateField))
	return r.GetAllPaginated(ctx, page, pageSize, opts...)
}

// ========== BatchRepository methods ==========

// CreateBatch inserts multiple entities
func (r *GormRepository[T]) CreateBatch(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&entities).Error
}

// UpdateBatch updates multiple entities
func (r *GormRepository[T]) UpdateBatch(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Save(entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple entities by IDs
func (r *GormRepository[T]) DeleteBatch(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, ids).Error
}

// ========== Advanced Query Helpers ==========

// Raw executes a raw SQL query and scans into the provided destination
func (r *GormRepository[T]) Raw(ctx context.Context, sql string, dest interface{}, args ...interface{}) error {
	return r.db.WithContext(ctx).Raw(sql, args...).Scan(dest).Error
}

// Exec executes a raw SQL statement
func (r *GormRepository[T]) Exec(ctx context.Context, sql string, args ...interface{}) error {
	return r.db.WithContext(ctx).Exec(sql, args...).Error
}

// WithTx creates a new repository instance using the provided transaction
func (r *GormRepository[T]) WithTx(tx *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: tx, dateField: r.dateField}
}

// Scopes applies custom query scopes
func (r *GormRepository[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{
		db:        r.db.Scopes(funcs...),
		dateField: r.dateField,
	}
}

// WhereRaw applies a raw WHERE clause
func (r *GormRepository[T]) WhereRaw(ctx context.Context, query string, args ...interface{}) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Where(query, args...).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// CustomQuery allows building a custom query and returning entities
func (r *GormRepository[T]) CustomQuery(ctx context.Context, buildQuery func(db *gorm.DB) *gorm.DB, opts ...QueryOption) ([]T, error) {
	var entities []T
	var sample T
	options := ApplyOptions(opts...)

	tx := r.db.WithContext(ctx)
	tx = buildQuery(tx)
	tx = r.applyOptions(tx, sample, options)

	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Preload returns a new repository with additional preloads for the next query
func (r *GormRepository[T]) Preload(relations ...string) *GormRepository[T] {
	tx := r.db
	for _, rel := range relations {
		tx = tx.Preload(rel)
	}
	return &GormRepository[T]{db: tx, dateField: r.dateField}
}

// Joins adds a JOIN clause for the next query
func (r *GormRepository[T]) Joins(query string, args ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{
		db:        r.db.Joins(query, args...),
		dateField: r.dateField,
	}
}

// Debug enables debug mode for the next query
func (r *GormRepository[T]) Debug() *GormRepository[T] {
	return &GormRepository[T]{
		db:        r.db.Debug(),
		dateField: r.dateField,
	}
}

// Verify interface implementation at compile time
var _ Repository[Entity] = (*GormRepository[Entity])(nil)

// Helper to format ordering with validation
func formatOrder(field string, desc bool) string {
	if field == "" {
		return ""
	}
	if desc {
		return fmt.Sprintf("%s DESC", field)
	}
	return fmt.Sprintf("%s ASC", field)
}
