package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"
	"fmt"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
)

func LoadExercisesOrderedForPlan(ctx context.Context, planID uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := database.From(ctx).Model(&models.Exercise{}).
		Joins("INNER JOIN workout_plan_exercises AS wpe ON wpe.exercise_id = exercises.id AND wpe.workout_plan_id = ?", planID).
		Order("wpe.display_order ASC").
		Find(&exercises).Error
	return exercises, err
}

func FindWorkoutPlanByID(ctx context.Context, planID uint) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := database.From(ctx).First(&plan, planID).Error
	return plan, err
}

func LoadPlanWithOrderedExercises(ctx context.Context, planID uint) (*models.WorkoutPlan, error) {
	plan, err := FindWorkoutPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	ex, err := LoadExercisesOrderedForPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	plan.Exercises = ex
	if err := loadAssignedDays(ctx, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func loadAssignedDays(ctx context.Context, plan *models.WorkoutPlan) error {
	var rows []models.WorkoutPlanDay
	if err := database.From(ctx).Where("workout_plan_id = ?", plan.ID).Order("day_of_week ASC").Find(&rows).Error; err != nil {
		return err
	}
	plan.AssignedDays = make([]int, 0, len(rows))
	for _, row := range rows {
		plan.AssignedDays = append(plan.AssignedDays, row.DayOfWeek)
	}
	if len(plan.AssignedDays) > 0 {
		day := plan.AssignedDays[0]
		plan.DayOfWeek = &day
	} else {
		plan.DayOfWeek = nil
	}
	return nil
}

func FindAllWorkoutPlans(ctx context.Context) ([]models.WorkoutPlan, error) {
	var workoutPlans []models.WorkoutPlan
	err := database.From(ctx).Find(&workoutPlans).Error
	if err != nil {
		return []models.WorkoutPlan{}, err
	}
	for i := range workoutPlans {
		loaded, err := LoadPlanWithOrderedExercises(ctx, workoutPlans[i].ID)
		if err != nil {
			return nil, err
		}
		workoutPlans[i] = *loaded
	}
	return workoutPlans, nil
}

func FindAllWorkoutPrograms(ctx context.Context) ([]models.WorkoutProgram, error) {
	var programs []models.WorkoutProgram
	if err := database.From(ctx).Order("id ASC").Find(&programs).Error; err != nil {
		return nil, err
	}
	for i := range programs {
		plans, err := FindWorkoutPlansByProgram(ctx, programs[i].ID)
		if err != nil {
			return nil, err
		}
		programs[i].Plans = plans
	}
	return programs, nil
}

func FindWorkoutPlansByProgram(ctx context.Context, programID uint) ([]models.WorkoutPlan, error) {
	var plans []models.WorkoutPlan
	if err := database.From(ctx).Where("workout_program_id = ?", programID).Order("id ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	for i := range plans {
		loaded, err := LoadPlanWithOrderedExercises(ctx, plans[i].ID)
		if err != nil {
			return nil, err
		}
		plans[i] = *loaded
	}
	return plans, nil
}

func CreateWorkoutProgram(ctx context.Context, program *models.WorkoutProgram) error {
	return database.From(ctx).Create(program).Error
}

func CreateWorkoutPlan(ctx context.Context, plan *models.WorkoutPlan) error {
	return database.From(ctx).Create(plan).Error
}

func FindWorkoutProgramByID(ctx context.Context, id uint) (models.WorkoutProgram, error) {
	var program models.WorkoutProgram
	return program, database.From(ctx).First(&program, id).Error
}

// FindActiveWorkoutProgram returns the active program. Older data can have several flagged active
// (the baseline "Default" program plus one created later): prefer the first one that has plans, the
// same rule the frontend's week strip uses, so a day created here gets the
// plan the UI shows for it.
func FindActiveWorkoutProgram(ctx context.Context) (models.WorkoutProgram, error) {
	var program models.WorkoutProgram
	err := database.From(ctx).Where("is_active = ?", true).
		Where("EXISTS (SELECT 1 FROM workout_plans wp WHERE wp.workout_program_id = workout_programs.id AND wp.deleted_at IS NULL)").
		Order("id ASC").First(&program).Error
	if err == gorm.ErrRecordNotFound {
		err = database.From(ctx).Where("is_active = ?", true).Order("id ASC").First(&program).Error
	}
	return program, err
}

func UpdateWorkoutProgramName(ctx context.Context, id uint, name string) error {
	return database.From(ctx).Model(&models.WorkoutProgram{}).Where("id = ?", id).Update("name", name).Error
}

func ActivateWorkoutProgram(ctx context.Context, id uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.WorkoutProgram{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return err
		}
		return tx.Model(&models.WorkoutProgram{}).Where("id != ?", id).Update("is_active", false).Error
	})
}

func WorkoutPlanExists(ctx context.Context, planID uint) (bool, error) {
	var n int64
	err := database.From(ctx).WithContext(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Count(&n).Error
	return n > 0, err
}

func FindWorkoutPlanByDayOfWeek(ctx context.Context, dayOfWeek int) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := database.From(ctx).Joins("INNER JOIN workout_plan_days AS wpd ON wpd.workout_plan_id = workout_plans.id").
		Where("wpd.day_of_week = ?", dayOfWeek).First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		err = database.From(ctx).Where("day_of_week = ?", dayOfWeek).First(&plan).Error
	}
	return plan, err
}

func FindWorkoutPlanByProgramAndDay(ctx context.Context, programID uint, dayOfWeek int) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := database.From(ctx).Joins("INNER JOIN workout_plan_days AS wpd ON wpd.workout_plan_id = workout_plans.id").
		Where("workout_plans.workout_program_id = ? AND wpd.day_of_week = ?", programID, dayOfWeek).
		First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		err = database.From(ctx).Where("workout_program_id = ? AND day_of_week = ?", programID, dayOfWeek).First(&plan).Error
	}
	return plan, err
}

func UpdatePlannedCardio(ctx context.Context, planID uint, cardioType string, minutes int) error {
	return database.From(ctx).Model(&models.WorkoutPlan{}).
		Where("id = ?", planID).
		Updates(map[string]any{
			"planned_cardio_type":    cardioType,
			"planned_cardio_minutes": minutes,
		}).Error
}

func UpdateMobilityItems(ctx context.Context, planID uint, preItems, postItems []string) error {
	var plan models.WorkoutPlan
	if err := database.From(ctx).First(&plan, planID).Error; err != nil {
		return err
	}
	plan.PreMobilityItems = preItems
	plan.PostMobilityItems = postItems
	return database.From(ctx).Model(&plan).
		Select("PreMobilityItems", "PostMobilityItems").
		Updates(&plan).Error
}

func AssignWorkoutPlanToProgram(ctx context.Context, planID uint, programID uint) error {
	return database.From(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("workout_program_id", programID).Error
}

func UnassignOtherPlansFromDay(ctx context.Context, dayOfWeek int, planID uint) error {
	return database.From(ctx).Model(&models.WorkoutPlan{}).
		Where("day_of_week = ? AND id != ?", dayOfWeek, planID).
		Update("day_of_week", nil).Error
}

func UnassignOtherPlansFromProgramDay(ctx context.Context, programID uint, dayOfWeek int, planID uint) error {
	var plans []models.WorkoutPlan
	if err := database.From(ctx).Where("workout_program_id = ? AND id != ?", programID, planID).Find(&plans).Error; err != nil {
		return err
	}
	for _, plan := range plans {
		if err := database.From(ctx).Where("workout_plan_id = ? AND day_of_week = ?", plan.ID, dayOfWeek).
			Delete(&models.WorkoutPlanDay{}).Error; err != nil {
			return err
		}
		if err := refreshLegacyDay(ctx, plan.ID); err != nil {
			return err
		}
	}
	return nil
}

func AssignWorkoutPlanToDay(ctx context.Context, planID uint, dayOfWeek int) error {
	if err := database.From(ctx).Where("workout_plan_id = ? AND day_of_week = ?", planID, dayOfWeek).
		FirstOrCreate(&models.WorkoutPlanDay{WorkoutPlanID: planID, DayOfWeek: dayOfWeek}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(ctx, planID)
}

func ClearWorkoutPlanDay(ctx context.Context, planID uint) error {
	if err := database.From(ctx).Where("workout_plan_id = ?", planID).Delete(&models.WorkoutPlanDay{}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(ctx, planID)
}

func ClearWorkoutPlanDayOfWeek(ctx context.Context, planID uint, dayOfWeek int) error {
	if err := database.From(ctx).Where("workout_plan_id = ? AND day_of_week = ?", planID, dayOfWeek).
		Delete(&models.WorkoutPlanDay{}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(ctx, planID)
}

func refreshLegacyDay(ctx context.Context, planID uint) error {
	var row models.WorkoutPlanDay
	err := database.From(ctx).Where("workout_plan_id = ?", planID).Order("day_of_week ASC").First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return database.From(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", nil).Error
	}
	if err != nil {
		return err
	}
	return database.From(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", row.DayOfWeek).Error
}

func AddExerciseToPlan(ctx context.Context, planID uint, exerciseID uint) error {
	if err := database.From(ctx).First(&models.WorkoutPlan{}, planID).Error; err != nil {
		return err
	}
	if err := database.From(ctx).First(&models.Exercise{}, exerciseID).Error; err != nil {
		return err
	}
	var n int64
	if err := database.From(ctx).Model(&models.WorkoutPlanExercise{}).Where("workout_plan_id = ? AND exercise_id = ?", planID, exerciseID).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("exercise already in plan")
	}
	var count int64
	if err := database.From(ctx).Model(&models.WorkoutPlanExercise{}).Where("workout_plan_id = ?", planID).Count(&count).Error; err != nil {
		return err
	}
	return database.From(ctx).Create(&models.WorkoutPlanExercise{
		WorkoutPlanID: planID,
		ExerciseID:    exerciseID,
		DisplayOrder:  int(count),
	}).Error
}

func renumberPlanExerciseDisplayOrder(ctx context.Context, planID uint) error {
	var rows []models.WorkoutPlanExercise
	if err := database.From(ctx).Where("workout_plan_id = ?", planID).Order("display_order ASC").Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		if rows[i].DisplayOrder != i {
			if err := database.From(ctx).Model(&models.WorkoutPlanExercise{}).
				Where("workout_plan_id = ? AND exercise_id = ?", planID, rows[i].ExerciseID).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveExerciseFromPlan(ctx context.Context, planID uint, exerciseID uint) error {
	res := database.From(ctx).Where("workout_plan_id = ? AND exercise_id = ?", planID, exerciseID).Delete(&models.WorkoutPlanExercise{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return renumberPlanExerciseDisplayOrder(ctx, planID)
}

func ReorderPlanExercises(ctx context.Context, planID uint, exerciseIDs []uint) error {
	var existing []models.WorkoutPlanExercise
	if err := database.From(ctx).Where("workout_plan_id = ?", planID).Find(&existing).Error; err != nil {
		return err
	}
	if len(exerciseIDs) != len(existing) {
		return fmt.Errorf("exercise list must include all plan exercises")
	}
	existingSet := make(map[uint]struct{}, len(existing))
	for _, e := range existing {
		existingSet[e.ExerciseID] = struct{}{}
	}
	for _, id := range exerciseIDs {
		if _, ok := existingSet[id]; !ok {
			return fmt.Errorf("invalid exercise id for plan")
		}
		delete(existingSet, id)
	}
	if len(existingSet) != 0 {
		return fmt.Errorf("exercise list must include all plan exercises")
	}
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		for i, eid := range exerciseIDs {
			if err := tx.Model(&models.WorkoutPlanExercise{}).
				Where("workout_plan_id = ? AND exercise_id = ?", planID, eid).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateWorkoutPlanName(ctx context.Context, id uint, name string) error {
	res := database.From(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", id).Update("name", name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SoftDeleteWorkoutPlan hides a routine but keeps its exercises and day rows, so
// RestoreWorkoutPlan can bring it back unchanged. Past logs keep their plan id.
func SoftDeleteWorkoutPlan(ctx context.Context, id uint) error {
	res := database.From(ctx).Delete(&models.WorkoutPlan{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// RestoreWorkoutPlan undoes SoftDeleteWorkoutPlan. Days that another live routine of the same
// program took in the meantime stay with that routine: this one drops them.
func RestoreWorkoutPlan(ctx context.Context, id uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		var plan models.WorkoutPlan
		if err := tx.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&plan).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&models.WorkoutPlan{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if plan.WorkoutProgramID == nil {
			return nil
		}
		taken := tx.Model(&models.WorkoutPlanDay{}).Select("workout_plan_days.day_of_week").
			Joins("INNER JOIN workout_plans AS wp ON wp.id = workout_plan_days.workout_plan_id").
			Where("wp.workout_program_id = ? AND wp.id != ? AND wp.deleted_at IS NULL", *plan.WorkoutProgramID, id)
		if err := tx.Where("workout_plan_id = ? AND day_of_week IN (?)", id, taken).
			Delete(&models.WorkoutPlanDay{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// RefreshLegacyDay re-syncs workout_plans.day_of_week with the plan's day rows.
func RefreshLegacyDay(ctx context.Context, planID uint) error { return refreshLegacyDay(ctx, planID) }
