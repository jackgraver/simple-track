package models

type FeatureModel interface {
	MigrateDatabase()
	seedDatabase() error
}

type Preloadable interface {
	Preloads() []string
}