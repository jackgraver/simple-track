package models

import "time"

// FeatureModel defines the interface for feature-level database setup
type FeatureModel interface {
	MigrateDatabase()
	seedDatabase() error
}

// Entity is the base interface all database models must implement
// This enables generic repository operations
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

// BaseModel provides common fields for all entities
// Embed this in your models for standard ID, timestamps
type BaseModel struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// GetID implements Entity interface
func (b BaseModel) GetID() uint {
	return b.ID
}