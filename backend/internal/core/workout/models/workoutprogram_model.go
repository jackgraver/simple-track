package models

import (
	"be-simpletracker/internal/database/repository"

	"gorm.io/gorm"
)

type WorkoutProgram struct {
	gorm.Model
	Name     string        `json:"name"`
	IsActive bool          `json:"is_active"`
	Plans    []WorkoutPlan `json:"plans" gorm:"foreignKey:WorkoutProgramID"`
}

func (w WorkoutProgram) GetID() uint       { return w.ID }
func (w WorkoutProgram) TableName() string { return "workout_programs" }

var _ repository.Entity = (*WorkoutProgram)(nil)
