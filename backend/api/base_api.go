package api

import "gorm.io/gorm"

type BaseFeature[T any] struct {
	db *gorm.DB
	SetEndpoints func()
}