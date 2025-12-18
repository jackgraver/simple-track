package repository

import "time"

// Entity is the base interface all database models must implement
type Entity interface {
	GetID() uint
	TableName() string
}

// Preloadable defines models that have eager-loading relationships
type Preloadable interface {
	// Preloads returns the default preload paths for this entity
	Preloads() []string
}

// Dateable defines models that have a date field for range queries
type Dateable interface {
	GetDate() time.Time
}

// SoftDeletable marks entities that use soft deletes
type SoftDeletable interface {
	IsDeleted() bool
}
