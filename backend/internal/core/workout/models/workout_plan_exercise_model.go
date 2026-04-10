package models

// WorkoutPlanExercise is the join row for plan ↔ exercise with display order.
type WorkoutPlanExercise struct {
	WorkoutPlanID uint `gorm:"primaryKey" json:"workout_plan_id"`
	ExerciseID    uint `gorm:"primaryKey" json:"exercise_id"`
	Position      int  `gorm:"not null;default:0" json:"position"`
}

func (WorkoutPlanExercise) TableName() string {
	return "workout_plan_exercises"
}
