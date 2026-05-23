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
	return &plan, nil
}

func FindAllWorkoutPlans() ([]models.WorkoutPlan, error) {
	var workoutPlans []models.WorkoutPlan
	err := conn().Find(&workoutPlans).Error
	if err != nil {
		return []models.WorkoutPlan{}, err
	}
	for i := range workoutPlans {
		ex, err := LoadExercisesOrderedForPlan(workoutPlans[i].ID)
		if err != nil {
			return nil, err
		}
		workoutPlans[i].Exercises = ex
	}
	return workoutPlans, nil
}

func WorkoutPlanExists(ctx context.Context, planID uint) (bool, error) {
	var n int64
	err := conn().WithContext(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", planID).Count(&n).Error
	return n > 0, err
}

func FindWorkoutPlanByDayOfWeek(dayOfWeek int) (models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := conn().Where("day_of_week = ?", dayOfWeek).First(&plan).Error
	return plan, err
}

func UpdatePlannedCardioType(planID uint, cardioType string) error {
	return conn().Model(&models.WorkoutPlan{}).
		Where("id = ?", planID).
		Update("planned_cardio_type", cardioType).Error
}

func UnassignOtherPlansFromDay(dayOfWeek int, planID uint) error {
	return conn().Model(&models.WorkoutPlan{}).
		Where("day_of_week = ? AND id != ?", dayOfWeek, planID).
		Update("day_of_week", nil).Error
}

func AssignWorkoutPlanToDay(planID uint, dayOfWeek int) error {
	dayOfWeekPtr := &dayOfWeek
	return conn().Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", dayOfWeekPtr).Error
}

func ClearWorkoutPlanDay(planID uint) error {
	return conn().Model(&models.WorkoutPlan{}).Where("id = ?", planID).Update("day_of_week", nil).Error
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
