package services

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/utils"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GetOrCreateToday(ctx context.Context, offset int) (models.WorkoutLog, error) {
	day := utils.ZerodTime(offset)
	workoutDay, err := workoutrepo.LoadByDate(ctx, day)
	if err == nil {
		return workoutDay, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.WorkoutLog{}, err
	}
	plan, err := GetPlanByDay(int(day.Weekday()))
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
	if err := workoutrepo.CreateMinimal(ctx, &newLog); err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutrepo.LoadByDate(ctx, day)
}

func SwitchPlan(ctx context.Context, offset int, planID *uint) (PreviousWorkoutResponse, error) {
	day, err := GetOrCreateToday(ctx, offset)
	if err != nil {
		return PreviousWorkoutResponse{}, err
	}
	if planID != nil {
		exists, err := workoutrepo.WorkoutPlanExists(ctx, *planID)
		if err != nil {
			return PreviousWorkoutResponse{}, err
		}
		if !exists {
			return PreviousWorkoutResponse{}, fmt.Errorf("workout plan not found")
		}
	}
	if err := workoutrepo.UpdateWorkoutPlanID(ctx, day.ID, planID); err != nil {
		return PreviousWorkoutResponse{}, err
	}
	return GetPreviousWorkoutView(ctx, offset)
}

type ExerciseGroup struct {
	Planned  *models.Exercise       `json:"planned,omitempty"`
	Logged   *models.LoggedExercise `json:"logged,omitempty"`
	Previous *models.LoggedExercise `json:"previous,omitempty"`
	Max      *models.LoggedExercise `json:"max,omitempty"`
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
	Day                 models.WorkoutLog    `json:"day"`
	PlannedExercises    []ExerciseGroup      `json:"planned_exercises"`
	PlannedCardio       any                  `json:"planned_cardio"`
	LoggedCardio        *models.Cardio       `json:"logged_cardio"`
	PlannedPreMobility  *MobilityRoutineView `json:"planned_pre_mobility"`
	LoggedPreMobility   *MobilityLoggedView  `json:"logged_pre_mobility"`
	PlannedPostMobility *MobilityRoutineView `json:"planned_post_mobility"`
	LoggedPostMobility  *MobilityLoggedView  `json:"logged_post_mobility"`
}

type MobilityRoutineView struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

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

func GetMonthWorkoutLogs(ctx context.Context, monthOffset int) (MonthWorkoutLogsResponse, error) {
	today := time.Now()
	target := today.AddDate(0, monthOffset, 0)
	startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
	end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))
	data, err := workoutrepo.GetByDateRange(ctx, start, end)
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

const (
	maxActivityRollingDays = 730
	maxActivityWeeks       = 104
)

var ErrInvalidActivityMode = errors.New("invalid activity mode")

type WorkoutActivityResponse struct {
	ActiveDates []string   `json:"active_dates"`
	Range       MonthRange `json:"range"`
	Mode        string     `json:"mode"`
}

func GetWorkoutActivity(ctx context.Context, mode string, weeks int) (WorkoutActivityResponse, error) {
	if mode != "year" && mode != "rolling" {
		return WorkoutActivityResponse{}, ErrInvalidActivityMode
	}
	now := time.Now()
	loc := now.Location()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	var start time.Time

	switch mode {
	case "year":
		y := now.Year()
		start = time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		end = time.Date(y, 12, 31, 0, 0, 0, 0, loc)
	case "rolling":
		w := weeks
		if w < 1 {
			w = 52
		}
		if w > maxActivityWeeks {
			w = maxActivityWeeks
		}
		start = end.AddDate(0, 0, -(w*7 - 1))
	}

	dates, err := workoutrepo.DatesWithLoggedSets(ctx, start, end)
	if err != nil {
		return WorkoutActivityResponse{}, err
	}
	active := make([]string, 0, len(dates))
	for _, d := range dates {
		d = d.In(loc)
		active = append(active, time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc).Format("2006-01-02"))
	}
	return WorkoutActivityResponse{
		ActiveDates: active,
		Range:       MonthRange{Start: start, End: end},
		Mode:        mode,
	}, nil
}

func UpsertCardio(ctx context.Context, offset int, minutes int, cardioType string, notes string) (*models.Cardio, error) {
	t, err := GetOrCreateToday(ctx, offset)
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
	existing, err := workoutrepo.FirstCardioByWorkoutLogID(ctx, t.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := models.Cardio{
			WorkoutLogID: t.ID,
			Minutes:      minutes,
			Type:         ctype,
			Notes:        notes,
		}
		if err := workoutrepo.CreateCardio(ctx, &row); err != nil {
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
	if err := workoutrepo.SaveCardio(ctx, &existing); err != nil {
		return nil, err
	}
	return &existing, nil
}

func UpsertCardioForWorkoutLog(ctx context.Context, offset int, minutes int, cardioType string, notes string) (*models.Cardio, error) {
	return UpsertCardio(ctx, offset, minutes, cardioType, notes)
}

func GetPreviousWorkoutView(ctx context.Context, offset int) (PreviousWorkoutResponse, error) {
	today, err := GetOrCreateToday(ctx, offset)
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
		prev, err := workoutrepo.GetPreviousExerciseLog(ctx, today.Date, p.Name, 0)
		if err == nil {
			group.Previous = &prev
		}
		maxLog, err := workoutrepo.GetMaxExerciseLog(ctx, today.Date, p.Name)
		if err == nil {
			group.Max = &maxLog
		}
		results = append(results, group)
	}
	for _, l := range loggedMap {
		if l.Exercise == nil {
			results = append(results, ExerciseGroup{Logged: &l})
			continue
		}
		group := ExerciseGroup{Logged: &l}
		prev, err := workoutrepo.GetPreviousExerciseLog(ctx, today.Date, l.Exercise.Name, 0)
		if err == nil {
			group.Previous = &prev
		}
		maxLog, err := workoutrepo.GetMaxExerciseLog(ctx, today.Date, l.Exercise.Name)
		if err == nil {
			group.Max = &maxLog
		}
		results = append(results, group)
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

func UpsertMobilityPre(ctx context.Context, offset int, checked []string) (*MobilityLoggedView, error) {
	t, err := GetOrCreateToday(ctx, offset)
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
	if err := workoutrepo.UpdatePreMobilityChecked(ctx, t.ID, filtered); err != nil {
		return nil, err
	}
	reloaded, err := workoutrepo.LoadByDate(ctx, utils.ZerodTime(offset))
	if err != nil {
		return nil, err
	}
	return loggedPreMobilityView(reloaded.WorkoutPlan, &reloaded), nil
}

func UpsertMobilityPost(ctx context.Context, offset int, checked []string) (*MobilityLoggedView, error) {
	t, err := GetOrCreateToday(ctx, offset)
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
	if err := workoutrepo.UpdatePostMobilityChecked(ctx, t.ID, filtered); err != nil {
		return nil, err
	}
	reloaded, err := workoutrepo.LoadByDate(ctx, utils.ZerodTime(offset))
	if err != nil {
		return nil, err
	}
	return loggedPostMobilityView(reloaded.WorkoutPlan, &reloaded), nil
}

func LogExercise(exercise *models.LoggedExercise) error {
	return workoutrepo.CreateLoggedExercise(exercise)
}

func UpdateLoggedExercise(exercise models.LoggedExercise) error {
	return workoutrepo.UpdateLoggedExerciseWithSets(exercise)
}

func RemoveLoggedExerciseForDay(ctx context.Context, offset int, exerciseID uint) error {
	return workoutrepo.RemoveLoggedExerciseForDay(ctx, utils.ZerodTime(offset), exerciseID)
}

func DeleteLoggedSet(setID uint) error {
	return workoutrepo.DeleteLoggedSet(setID)
}

func GetAllExercises(excludeIDs []uint) ([]models.Exercise, error) {
	return workoutrepo.FindAllExercises(excludeIDs)
}

type ExerciseProgressionEntry = workoutrepo.ExerciseProgressionEntry

func GetExerciseProgression(exerciseID uint) ([]ExerciseProgressionEntry, error) {
	return workoutrepo.GetExerciseProgression(exerciseID)
}

func LoadPlanWithOrderedExercises(planID uint) (*models.WorkoutPlan, error) {
	return workoutrepo.LoadPlanWithOrderedExercises(planID)
}

func GetAllWorkoutPlans() ([]models.WorkoutPlan, error) {
	return workoutrepo.FindAllWorkoutPlans()
}

func AddExerciseToPlan(planID uint, exerciseID uint) error {
	return workoutrepo.AddExerciseToPlan(planID, exerciseID)
}

func RemoveExerciseFromPlan(planID uint, exerciseID uint) error {
	return workoutrepo.RemoveExerciseFromPlan(planID, exerciseID)
}

func ReorderPlanExercises(planID uint, exerciseIDs []uint) error {
	return workoutrepo.ReorderPlanExercises(planID, exerciseIDs)
}

func CreateExercise(name string, repRollover uint, cues string) (*models.Exercise, error) {
	exercise := models.Exercise{
		Name:        name,
		RepRollover: repRollover,
		Cues:        cues,
	}
	if err := workoutrepo.CreateExercise(&exercise); err != nil {
		return nil, err
	}
	return &exercise, nil
}

func UpdateExercise(id uint, name string, repRollover uint, cues string) (*models.Exercise, error) {
	return workoutrepo.UpdateExercise(id, name, repRollover, cues)
}

func AssignPlanToDay(planID uint, dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}
	if _, err := workoutrepo.FindWorkoutPlanByID(planID); err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if err := workoutrepo.UnassignOtherPlansFromDay(dayOfWeek, planID); err != nil {
		return nil, fmt.Errorf("failed to unassign existing plan: %w", err)
	}
	if err := workoutrepo.AssignWorkoutPlanToDay(planID, dayOfWeek); err != nil {
		return nil, fmt.Errorf("failed to assign plan to day: %w", err)
	}
	reloaded, err := workoutrepo.LoadPlanWithOrderedExercises(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}
	return reloaded, nil
}

func UnassignPlanFromDay(planID uint) (*models.WorkoutPlan, error) {
	if _, err := workoutrepo.FindWorkoutPlanByID(planID); err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if err := workoutrepo.ClearWorkoutPlanDay(planID); err != nil {
		return nil, fmt.Errorf("failed to unassign plan from day: %w", err)
	}
	reloaded, err := workoutrepo.LoadPlanWithOrderedExercises(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}
	return reloaded, nil
}

func GetPlanByDay(dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}
	plan, err := workoutrepo.FindWorkoutPlanByDayOfWeek(dayOfWeek)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return workoutrepo.LoadPlanWithOrderedExercises(plan.ID)
}

func UpdateExerciseCues(exerciseID uint, cues string) (*models.Exercise, error) {
	return workoutrepo.UpdateExerciseCues(exerciseID, cues)
}

type ExerciseListResult = workoutrepo.ExerciseListResult

func ListExercises(page, pageSize int, search string) (ExerciseListResult, error) {
	return workoutrepo.ListExercises(page, pageSize, search)
}

func LoadLoggedExercise(id uint) (models.LoggedExercise, error) {
	return workoutrepo.LoadLoggedExercise(id)
}

func SetPlannedCardioType(planID uint, cardioType string) (*models.WorkoutPlan, error) {
	if err := workoutrepo.UpdatePlannedCardioType(planID, strings.TrimSpace(cardioType)); err != nil {
		return nil, err
	}
	return workoutrepo.LoadPlanWithOrderedExercises(planID)
}
