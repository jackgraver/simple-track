package workout

import (
	"be-simpletracker/internal/core/workout/models"
	"fmt"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.WorkoutProgram{},
		&models.Exercise{},
		&models.WorkoutPlan{},
		&models.WorkoutPlanDay{},
		&models.WorkoutPlanExercise{},
		&models.LoggedExercise{},
		&models.LoggedSet{},
		&models.WorkoutLog{},
		&models.Cardio{},
	); err != nil {
		return err
	}
	if db.Migrator().HasIndex(&models.WorkoutPlan{}, "idx_day_of_week") {
		if err := db.Migrator().DropIndex(&models.WorkoutPlan{}, "idx_day_of_week"); err != nil {
			return fmt.Errorf("drop legacy workout day index: %w", err)
		}
	}
	if db.Migrator().HasIndex(&models.WorkoutPlan{}, "idx_program_day_of_week") {
		if err := db.Migrator().DropIndex(&models.WorkoutPlan{}, "idx_program_day_of_week"); err != nil {
			return fmt.Errorf("drop legacy program day index: %w", err)
		}
	}
	var program models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&program).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.First(&program).Error; err == gorm.ErrRecordNotFound {
			program = models.WorkoutProgram{Name: "Default", IsActive: true}
			if err := db.Create(&program).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if err := db.Model(&program).Update("is_active", true).Error; err != nil {
			return err
		}
	}
	if err := db.Model(&models.WorkoutProgram{}).Where("id != ?", program.ID).Update("is_active", false).Error; err != nil {
		return err
	}
	if err := db.Model(&models.WorkoutPlan{}).
		Where("workout_program_id IS NULL").
		Update("workout_program_id", program.ID).Error; err != nil {
		return err
	}
	var plans []models.WorkoutPlan
	if err := db.Where("day_of_week IS NOT NULL").Find(&plans).Error; err != nil {
		return err
	}
	for _, plan := range plans {
		if err := db.Where("workout_plan_id = ? AND day_of_week = ?", plan.ID, *plan.DayOfWeek).
			FirstOrCreate(&models.WorkoutPlanDay{WorkoutPlanID: plan.ID, DayOfWeek: *plan.DayOfWeek}).Error; err != nil {
			return err
		}
	}
	return nil
}
