package models

import "time"

type WorkoutPlanDay struct {
	WorkoutPlanID uint      `json:"workout_plan_id" gorm:"primaryKey"`
	DayOfWeek     int       `json:"day_of_week" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"created_at"`
}

func (w WorkoutPlanDay) TableName() string {
	return "workout_plan_days"
}
