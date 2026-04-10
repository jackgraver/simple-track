package models

import "gorm.io/gorm"

// LoadExercisesOrderedForPlan returns exercises for a plan sorted by join position.
func LoadExercisesOrderedForPlan(db *gorm.DB, planID uint) ([]Exercise, error) {
	var exercises []Exercise
	err := db.Model(&Exercise{}).
		Joins("INNER JOIN workout_plan_exercises AS wpe ON wpe.exercise_id = exercises.id AND wpe.workout_plan_id = ?", planID).
		Order("wpe.position ASC").
		Find(&exercises).Error
	return exercises, err
}
