package workoutrepo

import (
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/database"
	dbrepo "be-simpletracker/internal/database/repository"

	"gorm.io/gorm"
)

func conn() *gorm.DB {
	return database.GetDB()
}

func LoadByDate(ctx context.Context, date time.Time) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog
	err := conn().WithContext(ctx).
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Preload("WorkoutPlan").
		Where("date = ?", date).
		First(&workoutDay).Error
	if err != nil {
		return models.WorkoutLog{}, err
	}
	if workoutDay.WorkoutPlan != nil {
		ex, err := models.LoadExercisesOrderedForPlan(conn(), workoutDay.WorkoutPlan.ID)
		if err != nil {
			return models.WorkoutLog{}, err
		}
		workoutDay.WorkoutPlan.Exercises = ex
	}
	return workoutDay, nil
}

func CreateMinimal(ctx context.Context, log *models.WorkoutLog) error {
	return conn().WithContext(ctx).Omit("WorkoutPlan", "Exercises", "Cardio").Create(log).Error
}

func GetByDateRange(ctx context.Context, start, end time.Time) ([]models.WorkoutLog, error) {
	repo := dbrepo.NewGormRepository[models.WorkoutLog](conn())
	return repo.GetByDateRange(ctx, start, end, dbrepo.WithDefaultPreloads())
}

func DatesWithLoggedSets(ctx context.Context, start, end time.Time) ([]time.Time, error) {
	var rows []struct {
		D time.Time `gorm:"column:d"`
	}
	err := conn().WithContext(ctx).Raw(`
SELECT workout_logs.date AS d
FROM logged_sets
JOIN logged_exercises ON logged_exercises.id = logged_sets.logged_exercise_id
JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id
WHERE workout_logs.date >= ? AND workout_logs.date <= ?
GROUP BY workout_logs.date
ORDER BY workout_logs.date ASC
`, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.D)
	}
	return out, nil
}

func GetPreviousExerciseLog(ctx context.Context, day time.Time, exercise string, offset int) (models.LoggedExercise, error) {
	var exerciseLog models.LoggedExercise
	err := conn().WithContext(ctx).
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Preload("Sets").
		Preload("Exercise").
		Order("workout_logs.date DESC").
		Offset(offset).
		Limit(1).
		Find(&exerciseLog).Error
	if err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

func FirstCardioByWorkoutLogID(ctx context.Context, workoutLogID uint) (models.Cardio, error) {
	var existing models.Cardio
	err := conn().WithContext(ctx).Where("workout_log_id = ?", workoutLogID).First(&existing).Error
	return existing, err
}

func CreateCardio(ctx context.Context, row *models.Cardio) error {
	return conn().WithContext(ctx).Create(row).Error
}

func SaveCardio(ctx context.Context, row *models.Cardio) error {
	return conn().WithContext(ctx).Save(row).Error
}

func UpdatePreMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := conn().WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PreMobilityChecked = checked
	return conn().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}

func UpdatePostMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := conn().WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PostMobilityChecked = checked
	return conn().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}
