package workoutrepo

import (
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"
	dbrepo "be-simpletracker/internal/database/repository"

	"gorm.io/gorm"
)

type WorkoutLogRepository struct {
	db *gorm.DB
}

func NewWorkoutLogRepository(db *gorm.DB) *WorkoutLogRepository {
	return &WorkoutLogRepository{db: db}
}

func (r *WorkoutLogRepository) LoadByDate(ctx context.Context, date time.Time) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog
	err := r.db.WithContext(ctx).
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Preload("WorkoutPlan.Exercises").
		Where("date = ?", date).
		First(&workoutDay).Error
	if err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

func (r *WorkoutLogRepository) CreateMinimal(ctx context.Context, log *models.WorkoutLog) error {
	return r.db.WithContext(ctx).Omit("WorkoutPlan", "Exercises", "Cardio").Create(log).Error
}

func (r *WorkoutLogRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]models.WorkoutLog, error) {
	repo := dbrepo.NewGormRepository[models.WorkoutLog](r.db)
	return repo.GetByDateRange(ctx, start, end, dbrepo.WithDefaultPreloads())
}

func (r *WorkoutLogRepository) GetPreviousExerciseLog(ctx context.Context, day time.Time, exercise string, offset int) (models.LoggedExercise, error) {
	var exerciseLog models.LoggedExercise
	err := r.db.WithContext(ctx).
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

func (r *WorkoutLogRepository) FirstCardioByWorkoutLogID(ctx context.Context, workoutLogID uint) (models.Cardio, error) {
	var existing models.Cardio
	err := r.db.WithContext(ctx).Where("workout_log_id = ?", workoutLogID).First(&existing).Error
	return existing, err
}

func (r *WorkoutLogRepository) CreateCardio(ctx context.Context, row *models.Cardio) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *WorkoutLogRepository) SaveCardio(ctx context.Context, row *models.Cardio) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *WorkoutLogRepository) UpdatePreMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := r.db.WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PreMobilityChecked = checked
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}

func (r *WorkoutLogRepository) UpdatePostMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := r.db.WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PostMobilityChecked = checked
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}
