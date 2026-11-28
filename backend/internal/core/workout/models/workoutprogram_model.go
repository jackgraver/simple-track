package models

import (
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

type WorkoutProgram struct {
	gorm.Model
	database.Owner
	Name     string        `json:"name"`
	IsActive bool          `json:"is_active"`
	Plans    []WorkoutPlan `json:"plans" gorm:"foreignKey:WorkoutProgramID"`
}

func (w WorkoutProgram) TableName() string { return "workout_programs" }
