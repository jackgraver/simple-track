package services

import (
	"be-simpletracker/internal/utils"
	"be-simpletracker/internal/workout/models"
	workoutrepo "be-simpletracker/internal/workout/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// WorkoutLogService coordinates workout log business logic and persistence.
type WorkoutLogService struct {
	db   *gorm.DB
	repo *workoutrepo.WorkoutLogRepository
}

func NewWorkoutLogService(db *gorm.DB) *WorkoutLogService {
	return &WorkoutLogService{
		db:   db,
		repo: workoutrepo.NewWorkoutLogRepository(db),
	}
}

// GetOrCreateToday returns the workout log for the calendar day (with offset from today),
// creating a row if missing and attaching the plan for that weekday when one exists.
func (s *WorkoutLogService) GetOrCreateToday(ctx context.Context, offset int) (models.WorkoutLog, error) {
	day := utils.ZerodTime(offset)
	workoutDay, err := s.repo.LoadByDate(ctx, day)
	if err == nil {
		return workoutDay, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.WorkoutLog{}, err
	}
	plan, err := GetPlanByDay(s.db, int(day.Weekday()))
	if err != nil {
		return models.WorkoutLog{}, err
	}
	var planID *uint
	if plan != nil {
		id := plan.ID
		planID = &id
	}
	newLog := models.WorkoutLog{
		Date:          day,
		WorkoutPlanID: planID,
	}
	if err := s.repo.CreateMinimal(ctx, &newLog); err != nil {
		return models.WorkoutLog{}, err
	}
	return s.repo.LoadByDate(ctx, day)
}

// GetOrCreateToday is a package-level helper for handlers that only have *gorm.DB.
func GetOrCreateToday(ctx context.Context, db *gorm.DB, offset int) (models.WorkoutLog, error) {
	return NewWorkoutLogService(db).GetOrCreateToday(ctx, offset)
}

type ExerciseGroup struct {
	Planned  *models.Exercise       `json:"planned,omitempty"`
	Logged   *models.LoggedExercise `json:"logged,omitempty"`
	Previous *models.LoggedExercise `json:"previous,omitempty"`
}

type MonthRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type MonthWorkoutLogsResponse struct {
	Days   []models.WorkoutLog `json:"days"`
	Today  time.Time           `json:"today"`
	Range  MonthRange          `json:"range"`
	Month  time.Month          `json:"month"`
	Offset int                 `json:"offset"`
}

type PreviousWorkoutResponse struct {
	Day                 models.WorkoutLog `json:"day"`
	PlannedExercises    []ExerciseGroup   `json:"planned_exercises"`
	PlannedCardio       any               `json:"planned_cardio"`
	LoggedCardio        *models.Cardio    `json:"logged_cardio"`
	PlannedPreMobility  *MobilityRoutineView `json:"planned_pre_mobility"`
	LoggedPreMobility   *MobilityLoggedView  `json:"logged_pre_mobility"`
	PlannedPostMobility *MobilityRoutineView `json:"planned_post_mobility"`
	LoggedPostMobility  *MobilityLoggedView  `json:"logged_post_mobility"`
}

// MobilityRoutineView is planned stretches for one slot (pre or post).
type MobilityRoutineView struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

// MobilityLoggedView is planned items plus which are checked for this workout log.
type MobilityLoggedView struct {
	Title   string   `json:"title"`
	Items   []string `json:"items"`
	Checked []string `json:"checked"`
}

const (
	preMobilityTitle  = "Pre-workout mobility"
	postMobilityTitle = "Post-workout mobility"
)

func plannedPreMobilityFromPlan(plan *models.WorkoutPlan) *MobilityRoutineView {
	if plan == nil || len(plan.PreMobilityItems) == 0 {
		return nil
	}
	return &MobilityRoutineView{Title: preMobilityTitle, Items: append([]string{}, plan.PreMobilityItems...)}
}

func plannedPostMobilityFromPlan(plan *models.WorkoutPlan) *MobilityRoutineView {
	if plan == nil || len(plan.PostMobilityItems) == 0 {
		return nil
	}
	return &MobilityRoutineView{Title: postMobilityTitle, Items: append([]string{}, plan.PostMobilityItems...)}
}

func filterCheckedToItems(items []string, checked []string) []string {
	valid := make(map[string]struct{}, len(items))
	for _, it := range items {
		valid[it] = struct{}{}
	}
	out := make([]string, 0, len(checked))
	for _, c := range checked {
		if _, ok := valid[c]; ok {
			out = append(out, c)
		}
	}
	return out
}

func loggedPreMobilityView(plan *models.WorkoutPlan, log *models.WorkoutLog) *MobilityLoggedView {
	var items []string
	if plan != nil {
		items = append([]string{}, plan.PreMobilityItems...)
	}
	checked := append([]string{}, log.PreMobilityChecked...)
	if len(items) == 0 && len(checked) == 0 {
		return nil
	}
	if len(items) == 0 {
		items = append([]string{}, checked...)
	}
	checked = filterCheckedToItems(items, checked)
	return &MobilityLoggedView{Title: preMobilityTitle, Items: items, Checked: checked}
}

func loggedPostMobilityView(plan *models.WorkoutPlan, log *models.WorkoutLog) *MobilityLoggedView {
	var items []string
	if plan != nil {
		items = append([]string{}, plan.PostMobilityItems...)
	}
	checked := append([]string{}, log.PostMobilityChecked...)
	if len(items) == 0 && len(checked) == 0 {
		return nil
	}
	if len(items) == 0 {
		items = append([]string{}, checked...)
	}
	checked = filterCheckedToItems(items, checked)
	return &MobilityLoggedView{Title: postMobilityTitle, Items: items, Checked: checked}
}

func plannedCardioFromPlan(plan *models.WorkoutPlan) any {
	if plan == nil {
		return nil
	}
	t := strings.TrimSpace(plan.PlannedCardioType)
	if t == "" {
		return nil
	}
	return map[string]any{"type": t}
}

func (s *WorkoutLogService) GetMonthWorkoutLogs(ctx context.Context, monthOffset int) (MonthWorkoutLogsResponse, error) {
	today := time.Now()
	target := today.AddDate(0, monthOffset, 0)
	startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
	end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))
	data, err := s.repo.GetByDateRange(ctx, start, end)
	if err != nil {
		return MonthWorkoutLogsResponse{}, err
	}
	return MonthWorkoutLogsResponse{
		Days:   data,
		Today:  today,
		Range:  MonthRange{Start: start, End: end},
		Month:  target.Month(),
		Offset: monthOffset,
	}, nil
}

func (s *WorkoutLogService) UpsertCardio(ctx context.Context, offset int, minutes int, cardioType string, notes string) (*models.Cardio, error) {
	t, err := s.GetOrCreateToday(ctx, offset)
	if err != nil {
		return nil, err
	}
	ctype := strings.TrimSpace(cardioType)
	if ctype == "" && t.WorkoutPlan != nil {
		ctype = strings.TrimSpace(t.WorkoutPlan.PlannedCardioType)
	}
	if ctype == "" {
		return nil, fmt.Errorf("cardio type is required when the plan has no planned cardio")
	}
	existing, err := s.repo.FirstCardioByWorkoutLogID(ctx, t.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := models.Cardio{
			WorkoutLogID: t.ID,
			Minutes:      minutes,
			Type:         ctype,
			Notes:        notes,
		}
		if err := s.repo.CreateCardio(ctx, &row); err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	existing.Minutes = minutes
	existing.Type = ctype
	existing.Notes = notes
	if err := s.repo.SaveCardio(ctx, &existing); err != nil {
		return nil, err
	}
	return &existing, nil
}

func UpsertCardioForWorkoutLog(ctx context.Context, db *gorm.DB, offset int, minutes int, cardioType string, notes string) (*models.Cardio, error) {
	return NewWorkoutLogService(db).UpsertCardio(ctx, offset, minutes, cardioType, notes)
}

func (s *WorkoutLogService) GetPreviousWorkoutView(ctx context.Context, offset int) (PreviousWorkoutResponse, error) {
	today, err := s.GetOrCreateToday(ctx, offset)
	if err != nil {
		return PreviousWorkoutResponse{}, err
	}
	logged := today.Exercises
	var planned []models.Exercise
	if today.WorkoutPlan != nil {
		planned = today.WorkoutPlan.Exercises
	}
	loggedMap := make(map[string]models.LoggedExercise)
	for _, l := range logged {
		if l.Exercise != nil {
			loggedMap[l.Exercise.Name] = l
		}
	}
	results := make([]ExerciseGroup, 0)
	for _, p := range planned {
		group := ExerciseGroup{Planned: &p}
		if log, ok := loggedMap[p.Name]; ok {
			group.Logged = &log
			delete(loggedMap, p.Name)
		}
		prev, err := s.repo.GetPreviousExerciseLog(ctx, today.Date, p.Name, 0)
		if err == nil {
			group.Previous = &prev
		}
		results = append(results, group)
	}
	for _, l := range loggedMap {
		if l.Exercise == nil {
			results = append(results, ExerciseGroup{Logged: &l})
			continue
		}
		prev, err := s.repo.GetPreviousExerciseLog(ctx, today.Date, l.Exercise.Name, 0)
		if err == nil {
			results = append(results, ExerciseGroup{
				Logged:   &l,
				Previous: &prev,
			})
		} else {
			results = append(results, ExerciseGroup{Logged: &l})
		}
	}
	return PreviousWorkoutResponse{
		Day:                 today,
		PlannedExercises:    results,
		PlannedCardio:       plannedCardioFromPlan(today.WorkoutPlan),
		LoggedCardio:        today.Cardio,
		PlannedPreMobility:  plannedPreMobilityFromPlan(today.WorkoutPlan),
		LoggedPreMobility:   loggedPreMobilityView(today.WorkoutPlan, &today),
		PlannedPostMobility: plannedPostMobilityFromPlan(today.WorkoutPlan),
		LoggedPostMobility:  loggedPostMobilityView(today.WorkoutPlan, &today),
	}, nil
}

func (s *WorkoutLogService) UpsertMobilityPre(ctx context.Context, offset int, checked []string) (*MobilityLoggedView, error) {
	t, err := s.GetOrCreateToday(ctx, offset)
	if err != nil {
		return nil, err
	}
	var items []string
	if t.WorkoutPlan != nil {
		items = append([]string{}, t.WorkoutPlan.PreMobilityItems...)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no pre-workout mobility planned for this day")
	}
	filtered := filterCheckedToItems(items, checked)
	if err := s.repo.UpdatePreMobilityChecked(ctx, t.ID, filtered); err != nil {
		return nil, err
	}
	reloaded, err := s.repo.LoadByDate(ctx, utils.ZerodTime(offset))
	if err != nil {
		return nil, err
	}
	return loggedPreMobilityView(reloaded.WorkoutPlan, &reloaded), nil
}

func (s *WorkoutLogService) UpsertMobilityPost(ctx context.Context, offset int, checked []string) (*MobilityLoggedView, error) {
	t, err := s.GetOrCreateToday(ctx, offset)
	if err != nil {
		return nil, err
	}
	var items []string
	if t.WorkoutPlan != nil {
		items = append([]string{}, t.WorkoutPlan.PostMobilityItems...)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no post-workout mobility planned for this day")
	}
	filtered := filterCheckedToItems(items, checked)
	if err := s.repo.UpdatePostMobilityChecked(ctx, t.ID, filtered); err != nil {
		return nil, err
	}
	reloaded, err := s.repo.LoadByDate(ctx, utils.ZerodTime(offset))
	if err != nil {
		return nil, err
	}
	return loggedPostMobilityView(reloaded.WorkoutPlan, &reloaded), nil
}

func LogExercise(db *gorm.DB, exercise *models.LoggedExercise) error {
	err := db.Omit("Exercise").Create(exercise).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateLoggedExercise(db *gorm.DB, exercise models.LoggedExercise) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.LoggedExercise{}).
			Where("id = ?", exercise.ID).
			Updates(map[string]any{
				"workout_log_id": exercise.WorkoutLogID,
				"exercise_id":    exercise.ExerciseID,
				"notes":          exercise.Notes,
			}).Error; err != nil {
			return err
		}

		incomingSetIDs := make(map[uint]struct{}, len(exercise.Sets))
		for i := range exercise.Sets {
			set := exercise.Sets[i]
			set.LoggedExerciseID = exercise.ID

			if set.ID > 0 {
				incomingSetIDs[set.ID] = struct{}{}
				if err := tx.Model(&models.LoggedSet{}).
					Where("id = ? AND logged_exercise_id = ?", set.ID, exercise.ID).
					Updates(map[string]any{
						"reps":               set.Reps,
						"weight":             set.Weight,
						"weight_setup":       set.WeightSetup,
						"logged_exercise_id": exercise.ID,
					}).Error; err != nil {
					return err
				}
				continue
			}

			set.ID = 0
			if err := tx.Create(&set).Error; err != nil {
				return err
			}
			incomingSetIDs[set.ID] = struct{}{}
		}

		var existingSetIDs []uint
		if err := tx.Model(&models.LoggedSet{}).
			Where("logged_exercise_id = ?", exercise.ID).
			Pluck("id", &existingSetIDs).Error; err != nil {
			return err
		}

		setIDsToDelete := make([]uint, 0)
		for _, existingSetID := range existingSetIDs {
			if _, ok := incomingSetIDs[existingSetID]; !ok {
				setIDsToDelete = append(setIDsToDelete, existingSetID)
			}
		}

		if len(setIDsToDelete) > 0 {
			if err := tx.Unscoped().
				Where("logged_exercise_id = ? AND id IN ?", exercise.ID, setIDsToDelete).
				Delete(&models.LoggedSet{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func DeleteLoggedSet(db *gorm.DB, setID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var set models.LoggedSet
		if err := tx.Where("id = ?", setID).First(&set).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Delete(&set).Error; err != nil {
			return err
		}

		var remainingSetCount int64
		if err := tx.Model(&models.LoggedSet{}).
			Where("logged_exercise_id = ?", set.LoggedExerciseID).
			Count(&remainingSetCount).Error; err != nil {
			return err
		}

		if remainingSetCount == 0 {
			if err := tx.Unscoped().
				Delete(&models.LoggedExercise{}, set.LoggedExerciseID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func GetAllExercises(db *gorm.DB, excludeIDs []uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	query := db.Model(&models.Exercise{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Find(&exercises).Error
	if err != nil {
		return []models.Exercise{}, err
	}
	return exercises, nil
}

type ExerciseProgressionEntry struct {
	Date   time.Time `json:"date"`
	Weight float32   `json:"weight"`
	Reps   uint      `json:"reps"`
}

func GetExerciseProgression(db *gorm.DB, exerciseID uint) ([]ExerciseProgressionEntry, error) {
	var entries []ExerciseProgressionEntry

	err := db.
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

func GetAllWorkoutPlans(db *gorm.DB) ([]models.WorkoutPlan, error) {
	var workoutPlans []models.WorkoutPlan
	err := db.Preload("Exercises").Find(&workoutPlans).Error
	if err != nil {
		return []models.WorkoutPlan{}, err
	}
	return workoutPlans, nil
}

func AddExerciseToPlan(db *gorm.DB, planID uint, exerciseID uint) error {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return err
	}

	var exercise models.Exercise
	if err := db.First(&exercise, exerciseID).Error; err != nil {
		return err
	}

	return db.Model(&plan).Association("Exercises").Append(&exercise)
}

func RemoveExerciseFromPlan(db *gorm.DB, planID uint, exerciseID uint) error {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return err
	}

	var exercise models.Exercise
	if err := db.First(&exercise, exerciseID).Error; err != nil {
		return err
	}

	return db.Model(&plan).Association("Exercises").Delete(&exercise)
}

func CreateExercise(db *gorm.DB, name string, repRollover uint, cues string) (*models.Exercise, error) {
	exercise := models.Exercise{
		Name:        name,
		RepRollover: repRollover,
		Cues:        cues,
	}

	if err := db.Create(&exercise).Error; err != nil {
		return nil, err
	}

	return &exercise, nil
}

// AssignPlanToDay assigns a workout plan to a specific day of the week
// If another plan is already assigned to that day, it will be unassigned first
// dayOfWeek: 0=Sunday, 1=Monday, ..., 6=Saturday
func AssignPlanToDay(db *gorm.DB, planID uint, dayOfWeek int) (*models.WorkoutPlan, error) {
	// Validate dayOfWeek
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}

	// Find the plan
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Unassign any existing plan from this day
	if err := db.Model(&models.WorkoutPlan{}).
		Where("day_of_week = ? AND id != ?", dayOfWeek, planID).
		Update("day_of_week", nil).Error; err != nil {
		return nil, fmt.Errorf("failed to unassign existing plan: %w", err)
	}

	// Assign the plan to the day
	dayOfWeekPtr := &dayOfWeek
	if err := db.Model(&plan).Update("day_of_week", dayOfWeekPtr).Error; err != nil {
		return nil, fmt.Errorf("failed to assign plan to day: %w", err)
	}

	// Reload the plan with exercises
	if err := db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}

	return &plan, nil
}

// UnassignPlanFromDay removes the day assignment from a workout plan
func UnassignPlanFromDay(db *gorm.DB, planID uint) (*models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	if err := db.Model(&plan).Update("day_of_week", nil).Error; err != nil {
		return nil, fmt.Errorf("failed to unassign plan from day: %w", err)
	}

	// Reload the plan with exercises
	if err := db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}

	return &plan, nil
}

// GetPlanByDay returns the workout plan assigned to a specific day of the week.
func GetPlanByDay(db *gorm.DB, dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}

	var plan models.WorkoutPlan
	err := db.Preload("Exercises").Where("day_of_week = ?", dayOfWeek).First(&plan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No plan assigned to this day
		}
		return nil, err
	}

	return &plan, nil
}
