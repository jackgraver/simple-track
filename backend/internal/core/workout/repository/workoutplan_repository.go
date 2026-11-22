package workoutrepo

import (
	"context"
	"fmt"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
)

func LoadExercisesOrderedForPlan(planID uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := conn().Model(&models.Exercise{}).
		Joins("INNER JOIN workout_plan_exercises AS wpe ON wpe.exercise_id = exercises.id AND wpe.workout_plan_id = ?", planID).
		Order("wpe.display_order ASC").
		Find(&exercises).Error
	return exercises, err
}

func FindWorkoutPlanByID(planID uint) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := conn().First(&plan, planID).Error
	return plan, err
}

func LoadPlanWithOrderedExercises(planID uint) (*models.WorkoutPlan, error) {
	plan, err := FindWorkoutPlanByID(planID)
	if err != nil {
		return nil, err
	}
	ex, err := LoadExercisesOrderedForPlan(planID)
	if err != nil {
		return nil, err
	}
	plan.Exercises = ex
	if err := loadAssignedDays(&plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func loadAssignedDays(plan *models.WorkoutPlan) error {
	var rows []models.WorkoutPlanDay
	if err := conn().Where("workout_plan_id = ?", plan.ID).Order("day_of_week ASC").Find(&rows).Error; err != nil {
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

func FindAllWorkoutPlans() ([]models.WorkoutPlan, error) {
	var workoutPlans []models.WorkoutPlan
	err := conn().Find(&workoutPlans).Error
	if err != nil {
		return []models.WorkoutPlan{}, err
	}
	for i := range workoutPlans {
		loaded, err := LoadPlanWithOrderedExercises(workoutPlans[i].ID)
		if err != nil {
			return nil, err
		}
		workoutPlans[i] = *loaded
	}
	return workoutPlans, nil
}

func FindAllWorkoutPrograms() ([]models.WorkoutProgram, error) {
	var programs []models.WorkoutProgram
	if err := conn().Order("id ASC").Find(&programs).Error; err != nil {
		return nil, err
	}
	for i := range programs {
		plans, err := FindWorkoutPlansByProgram(programs[i].ID)
		if err != nil {
			return nil, err
		}
		programs[i].Plans = plans
	}
	return programs, nil
}

func FindWorkoutPlansByProgram(programID uint) ([]models.WorkoutPlan, error) {
	var plans []models.WorkoutPlan
	if err := conn().Where("workout_program_id = ?", programID).Order("id ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	for i := range plans {
		loaded, err := LoadPlanWithOrderedExercises(plans[i].ID)
		if err != nil {
			return nil, err
		}
		plans[i] = *loaded
	}
	return plans, nil
}

func CreateWorkoutProgram(program *models.WorkoutProgram) error {
	return conn().Create(program).Error
}

func CreateWorkoutPlan(plan *models.WorkoutPlan) error {
	return conn().Create(plan).Error
}

func FindWorkoutProgramByID(id uint) (models.WorkoutProgram, error) {
	var program models.WorkoutProgram
	return program, conn().First(&program, id).Error
}

// FindActiveWorkoutProgram returns the active program. Older data can have several flagged active
// (the baseline "Default" program plus one created later): prefer the first one that has plans, the
// same rule the frontend's week strip uses, so a day created here gets the
// plan the UI shows for it.
func FindActiveWorkoutProgram() (models.WorkoutProgram, error) {
	var program models.WorkoutProgram
	err := conn().Where("is_active = ?", true).
		Where("EXISTS (SELECT 1 FROM workout_plans wp WHERE wp.workout_program_id = workout_programs.id AND wp.deleted_at IS NULL)").
		Order("id ASC").First(&program).Error
	if err == gorm.ErrRecordNotFound {
		err = conn().Where("is_active = ?", true).Order("id ASC").First(&program).Error
	}
	return program, err
}

func UpdateWorkoutProgramName(id uint, name string) error {
	return conn().Model(&models.WorkoutProgram{}).Where("id = ?", id).Update("name", name).Error
}

func ActivateWorkoutProgram(id uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.WorkoutProgram{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return err
		}
		return tx.Model(&models.WorkoutProgram{}).Where("id != ?", id).Update("is_active", false).Error
	})
}

func WorkoutPlanExists(ctx context.Context, planID uint) (bool, error) {
	var n int64
	err := conn().WithContext(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Count(&n).Error
	return n > 0, err
}

func FindWorkoutPlanByDayOfWeek(dayOfWeek int) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := conn().Joins("INNER JOIN workout_plan_days AS wpd ON wpd.workout_plan_id = workout_plans.id").
		Where("wpd.day_of_week = ?", dayOfWeek).First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		err = conn().Where("day_of_week = ?", dayOfWeek).First(&plan).Error
	}
	return plan, err
}

func FindWorkoutPlanByProgramAndDay(programID uint, dayOfWeek int) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := conn().Joins("INNER JOIN workout_plan_days AS wpd ON wpd.workout_plan_id = workout_plans.id").
		Where("workout_plans.workout_program_id = ? AND wpd.day_of_week = ?", programID, dayOfWeek).
		First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		err = conn().Where("workout_program_id = ? AND day_of_week = ?", programID, dayOfWeek).First(&plan).Error
	}
	return plan, err
}

func UpdatePlannedCardio(planID uint, cardioType string, minutes int) error {
	return conn().Model(&models.WorkoutPlan{}).
		Where("id = ?", planID).
		Updates(map[string]any{
			"planned_cardio_type":    cardioType,
			"planned_cardio_minutes": minutes,
		}).Error
}

func UpdateMobilityItems(planID uint, preItems, postItems []string) error {
	var plan models.WorkoutPlan
	if err := conn().First(&plan, planID).Error; err != nil {
		return err
	}
	plan.PreMobilityItems = preItems
	plan.PostMobilityItems = postItems
	return conn().Model(&plan).
		Select("PreMobilityItems", "PostMobilityItems").
		Updates(&plan).Error
}

func AssignWorkoutPlanToProgram(planID uint, programID uint) error {
	return conn().Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("workout_program_id", programID).Error
}

func UnassignOtherPlansFromDay(dayOfWeek int, planID uint) error {
	return conn().Model(&models.WorkoutPlan{}).
		Where("day_of_week = ? AND id != ?", dayOfWeek, planID).
		Update("day_of_week", nil).Error
}

func UnassignOtherPlansFromProgramDay(programID uint, dayOfWeek int, planID uint) error {
	var plans []models.WorkoutPlan
	if err := conn().Where("workout_program_id = ? AND id != ?", programID, planID).Find(&plans).Error; err != nil {
		return err
	}
	for _, plan := range plans {
		if err := conn().Where("workout_plan_id = ? AND day_of_week = ?", plan.ID, dayOfWeek).
			Delete(&models.WorkoutPlanDay{}).Error; err != nil {
			return err
		}
		if err := refreshLegacyDay(plan.ID); err != nil {
			return err
		}
	}
	return nil
}

func AssignWorkoutPlanToDay(planID uint, dayOfWeek int) error {
	if err := conn().Where("workout_plan_id = ? AND day_of_week = ?", planID, dayOfWeek).
		FirstOrCreate(&models.WorkoutPlanDay{WorkoutPlanID: planID, DayOfWeek: dayOfWeek}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(planID)
}

func ClearWorkoutPlanDay(planID uint) error {
	if err := conn().Where("workout_plan_id = ?", planID).Delete(&models.WorkoutPlanDay{}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(planID)
}

func ClearWorkoutPlanDayOfWeek(planID uint, dayOfWeek int) error {
	if err := conn().Where("workout_plan_id = ? AND day_of_week = ?", planID, dayOfWeek).
		Delete(&models.WorkoutPlanDay{}).Error; err != nil {
		return err
	}
	return refreshLegacyDay(planID)
}

func refreshLegacyDay(planID uint) error {
	var row models.WorkoutPlanDay
	err := conn().Where("workout_plan_id = ?", planID).Order("day_of_week ASC").First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return conn().Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", nil).Error
	}
	if err != nil {
		return err
	}
	return conn().Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", row.DayOfWeek).Error
}

func AddExerciseToPlan(planID uint, exerciseID uint) error {
	if err := conn().First(&models.WorkoutPlan{}, planID).Error; err != nil {
		return err
	}
	if err := conn().First(&models.Exercise{}, exerciseID).Error; err != nil {
		return err
	}
	var n int64
	if err := conn().Model(&models.WorkoutPlanExercise{}).Where("workout_plan_id = ? AND exercise_id = ?", planID, exerciseID).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("exercise already in plan")
	}
	var count int64
	if err := conn().Model(&models.WorkoutPlanExercise{}).Where("workout_plan_id = ?", planID).Count(&count).Error; err != nil {
		return err
	}
	return conn().Create(&models.WorkoutPlanExercise{
		WorkoutPlanID: planID,
		ExerciseID:    exerciseID,
		DisplayOrder:  int(count),
	}).Error
}

func renumberPlanExerciseDisplayOrder(planID uint) error {
	var rows []models.WorkoutPlanExercise
	if err := conn().Where("workout_plan_id = ?", planID).Order("display_order ASC").Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		if rows[i].DisplayOrder != i {
			if err := conn().Model(&models.WorkoutPlanExercise{}).
				Where("workout_plan_id = ? AND exercise_id = ?", planID, rows[i].ExerciseID).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveExerciseFromPlan(planID uint, exerciseID uint) error {
	res := conn().Where("workout_plan_id = ? AND exercise_id = ?", planID, exerciseID).Delete(&models.WorkoutPlanExercise{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return renumberPlanExerciseDisplayOrder(planID)
}

func ReorderPlanExercises(planID uint, exerciseIDs []uint) error {
	var existing []models.WorkoutPlanExercise
	if err := conn().Where("workout_plan_id = ?", planID).Find(&existing).Error; err != nil {
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
	return conn().Transaction(func(tx *gorm.DB) error {
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

func UpdateWorkoutPlanName(id uint, name string) error {
	res := conn().Model(&models.WorkoutPlan{}).Where("id = ?", id).Update("name", name)
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
func SoftDeleteWorkoutPlan(id uint) error {
	res := conn().Delete(&models.WorkoutPlan{}, id)
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
func RestoreWorkoutPlan(id uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
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
func RefreshLegacyDay(planID uint) error { return refreshLegacyDay(planID) }
