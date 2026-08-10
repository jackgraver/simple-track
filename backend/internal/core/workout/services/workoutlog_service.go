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
	preMobilityTitle  = "Dynamic warmup"
	postMobilityTitle = "Static stretching"
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
	return map[string]any{
		"type":    t,
		"minutes": plan.PlannedCardioMinutes,
	}
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
		return nil, fmt.Errorf("no dynamic warmup planned for this day")
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
		return nil, fmt.Errorf("no static stretching planned for this day")
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

func GetAllWorkoutPrograms() ([]models.WorkoutProgram, error) {
	return workoutrepo.FindAllWorkoutPrograms()
}

func CreateWorkoutProgram(name string) (*models.WorkoutProgram, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("program name is required")
	}
	program := &models.WorkoutProgram{Name: name}
	if _, err := workoutrepo.FindActiveWorkoutProgram(); err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		program.IsActive = true
	}
	if err := workoutrepo.CreateWorkoutProgram(program); err != nil {
		return nil, err
	}
	return program, nil
}

func CreateWorkoutPlan(programID uint, name string, dayOfWeek *int) (*models.WorkoutPlan, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("plan name is required")
	}
	if _, err := workoutrepo.FindWorkoutProgramByID(programID); err != nil {
		return nil, fmt.Errorf("program not found: %w", err)
	}
	if dayOfWeek != nil && (*dayOfWeek < 0 || *dayOfWeek > 6) {
		return nil, fmt.Errorf("day_of_week must be between 0 and 6")
	}
	plan := &models.WorkoutPlan{Name: name, WorkoutProgramID: &programID}
	if err := workoutrepo.CreateWorkoutPlan(plan); err != nil {
		return nil, err
	}
	if dayOfWeek != nil {
		if _, err := AssignPlanToDay(plan.ID, *dayOfWeek); err != nil {
			return nil, err
		}
	}
	return workoutrepo.LoadPlanWithOrderedExercises(plan.ID)
}

func RenameWorkoutProgram(id uint, name string) (*models.WorkoutProgram, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("program name is required")
	}
	if err := workoutrepo.UpdateWorkoutProgramName(id, name); err != nil {
		return nil, err
	}
	program, err := workoutrepo.FindWorkoutProgramByID(id)
	if err != nil {
		return nil, err
	}
	return &program, nil
}

func ActivateWorkoutProgram(ctx context.Context, id uint) (*models.WorkoutProgram, error) {
	if _, err := workoutrepo.FindWorkoutProgramByID(id); err != nil {
		return nil, fmt.Errorf("program not found: %w", err)
	}
	if _, err := GetOrCreateToday(ctx, 0); err != nil {
		return nil, err
	}
	if err := workoutrepo.ActivateWorkoutProgram(id); err != nil {
		return nil, err
	}
	program, err := workoutrepo.FindWorkoutProgramByID(id)
	if err != nil {
		return nil, err
	}
	return &program, nil
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

func CreateExercise(name string, repRollover uint, cues string, loadTypes ...models.ExerciseLoadType) (*models.Exercise, error) {
	loadType := models.ExerciseLoadType("")
	if len(loadTypes) > 0 {
		loadType = loadTypes[0]
	}
	exercise := models.Exercise{
		Name:        name,
		RepRollover: repRollover,
		Cues:        cues,
		LoadType:    models.NormalizeExerciseLoadType(loadType),
	}
	if err := workoutrepo.CreateExercise(&exercise); err != nil {
		return nil, err
	}
	return &exercise, nil
}

func UpdateExercise(id uint, name string, repRollover uint, cues string, loadTypes ...models.ExerciseLoadType) (*models.Exercise, error) {
	return workoutrepo.UpdateExercise(id, name, repRollover, cues, loadTypes...)
}

func AssignPlanToDay(planID uint, dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}
	plan, err := workoutrepo.FindWorkoutPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if plan.WorkoutProgramID == nil {
		program, err := workoutrepo.FindActiveWorkoutProgram()
		if err != nil {
			return nil, fmt.Errorf("active program not found: %w", err)
		}
		if err := workoutrepo.AssignWorkoutPlanToProgram(planID, program.ID); err != nil {
			return nil, err
		}
		plan.WorkoutProgramID = &program.ID
	}
	if err := workoutrepo.UnassignOtherPlansFromProgramDay(*plan.WorkoutProgramID, dayOfWeek, planID); err != nil {
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
	return unassignPlanFromDay(planID, nil)
}

func UnassignPlanFromSpecificDay(planID uint, dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}
	return unassignPlanFromDay(planID, &dayOfWeek)
}

func unassignPlanFromDay(planID uint, dayOfWeek *int) (*models.WorkoutPlan, error) {
	if _, err := workoutrepo.FindWorkoutPlanByID(planID); err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	var err error
	if dayOfWeek == nil {
		err = workoutrepo.ClearWorkoutPlanDay(planID)
	} else {
		err = workoutrepo.ClearWorkoutPlanDayOfWeek(planID, *dayOfWeek)
	}
	if err != nil {
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
	program, err := workoutrepo.FindActiveWorkoutProgram()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	plan, err := workoutrepo.FindWorkoutPlanByProgramAndDay(program.ID, dayOfWeek)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			legacy, legacyErr := workoutrepo.FindWorkoutPlanByDayOfWeek(dayOfWeek)
			if legacyErr == gorm.ErrRecordNotFound {
				return nil, nil
			}
			if legacyErr != nil {
				return nil, legacyErr
			}
			if legacy.WorkoutProgramID == nil {
				if err := workoutrepo.AssignWorkoutPlanToProgram(legacy.ID, program.ID); err != nil {
					return nil, err
				}
			}
			plan = legacy
		} else {
			return nil, err
		}
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

func SetPlannedCardio(planID uint, cardioType string, minutes int) (*models.WorkoutPlan, error) {
	if minutes < 0 {
		return nil, errors.New("planned cardio minutes cannot be negative")
	}
	cardioType = strings.TrimSpace(cardioType)
	if cardioType == "" {
		minutes = 0
	}
	if err := workoutrepo.UpdatePlannedCardio(planID, cardioType, minutes); err != nil {
		return nil, err
	}
	return workoutrepo.LoadPlanWithOrderedExercises(planID)
}

func SetPlannedMobility(planID uint, preItems, postItems []string) (*models.WorkoutPlan, error) {
	normalize := func(items []string) []string {
		result := make([]string, 0, len(items))
		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			key := strings.ToLower(item)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, item)
		}
		return result
	}
	preItems = normalize(preItems)
	postItems = normalize(postItems)
	if err := workoutrepo.UpdateMobilityItems(planID, preItems, postItems); err != nil {
		return nil, err
	}
	return workoutrepo.LoadPlanWithOrderedExercises(planID)
}
