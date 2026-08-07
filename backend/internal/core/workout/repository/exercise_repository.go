package workoutrepo

import (
	"time"

	"be-simpletracker/internal/core/workout/models"
)

type ExerciseListResult struct {
	Exercises []models.Exercise
	Total     int64
}

type ExerciseProgressionEntry struct {
	Date   time.Time `json:"date"`
	Weight float32   `json:"weight"`
	Reps   uint      `json:"reps"`
}

func FindAllExercises(excludeIDs []uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	query := conn().Model(&models.Exercise{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Find(&exercises).Error
	if err != nil {
		return []models.Exercise{}, err
	}
	return exercises, nil
}

func ListExercises(page, pageSize int, search string) (ExerciseListResult, error) {
	query := conn().Model(&models.Exercise{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	query = query.Order("name ASC")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ExerciseListResult{}, err
	}

	var exercises []models.Exercise
	if pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	if err := query.Find(&exercises).Error; err != nil {
		return ExerciseListResult{}, err
	}

	return ExerciseListResult{Exercises: exercises, Total: total}, nil
}

func GetExerciseProgression(exerciseID uint) ([]ExerciseProgressionEntry, error) {
	var entries []ExerciseProgressionEntry

	err := conn().
		Table("logged_exercises").
		Select("workout_logs.date, logged_sets.weight, logged_sets.reps").
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN logged_sets ON logged_sets.logged_exercise_id = logged_exercises.id").
		Where("logged_exercises.exercise_id = ?", exerciseID).
		Where("logged_sets.weight > 0 AND logged_sets.reps > 0").
		Order("workout_logs.date ASC").
		Scan(&entries).Error

	if err != nil {
		return []ExerciseProgressionEntry{}, err
	}
	return entries, nil
}

func ExerciseExists(id uint) error {
	return conn().First(&models.Exercise{}, id).Error
}

func CreateExercise(exercise *models.Exercise) error {
	return conn().Create(exercise).Error
}

func UpdateExercise(id uint, name string, repRollover uint, cues string, loadTypes ...models.ExerciseLoadType) (*models.Exercise, error) {
	var exercise models.Exercise
	if err := conn().First(&exercise, id).Error; err != nil {
		return nil, err
	}
	loadType := models.NormalizeExerciseLoadType(exercise.LoadType)
	if len(loadTypes) > 0 {
		loadType = models.NormalizeExerciseLoadType(loadTypes[0])
	}
	if err := conn().Model(&exercise).Updates(map[string]interface{}{
		"name":         name,
		"rep_rollover": repRollover,
		"cues":         cues,
		"load_type":    loadType,
	}).Error; err != nil {
		return nil, err
	}
	if err := conn().First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

func UpdateExerciseCues(exerciseID uint, cues string) (*models.Exercise, error) {
	var exercise models.Exercise
	if err := conn().First(&exercise, exerciseID).Error; err != nil {
		return nil, err
	}
	if err := conn().Model(&exercise).Update("cues", cues).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}
