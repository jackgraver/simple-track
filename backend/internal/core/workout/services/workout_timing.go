package services

import (
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"context"
	"sort"
	"time"
)

// Timing stats are derived from logged_sets.created_at, which is stamped when
// a set is saved (right after it is performed). Nothing records when a set
// starts, so:
//   - workout duration = first set saved -> last set saved
//   - set interval     = gap between consecutive sets of the same exercise
//     (work + rest)
//   - transition       = gap between the last set of anything else and the
//     first set of an exercise (setup + first set)
const (
	// Sets saved outside [log date, log date + window) were logged after the
	// fact (e.g. backfilling a past day), so their timestamps mean nothing.
	timingSameDayWindow = 36 * time.Hour
	// A session whose median gap is below this was bulk-entered, not logged live.
	timingBulkEntryMedianGap = 15 * time.Second
	timingMinSetGap          = 5 * time.Second
	timingMaxSetGap          = 20 * time.Minute
	timingMaxTransitionGap   = 30 * time.Minute
	timingMaxWorkout         = 4 * time.Hour

	defaultTimingDays = 90
	maxTimingDays     = 730
)

type DurationStat struct {
	Count         int     `json:"count"`
	AvgSeconds    float64 `json:"avg_seconds"`
	MedianSeconds float64 `json:"median_seconds"`
}

type ExerciseTimingStat struct {
	ExerciseID  uint         `json:"exercise_id"`
	Name        string       `json:"name"`
	SetInterval DurationStat `json:"set_interval"`
	Transition  DurationStat `json:"transition"`
}

type WorkoutTimingStats struct {
	Days             int                  `json:"days"`
	Start            time.Time            `json:"start"`
	Workout          DurationStat         `json:"workout"`
	SetInterval      DurationStat         `json:"set_interval"`
	Transition       DurationStat         `json:"transition"`
	Exercises        []ExerciseTimingStat `json:"exercises"`
	ExcludedWorkouts int                  `json:"excluded_workouts"`
}

func GetWorkoutTimingStats(ctx context.Context, days int) (WorkoutTimingStats, error) {
	if days <= 0 {
		days = defaultTimingDays
	}
	if days > maxTimingDays {
		days = maxTimingDays
	}
	start := zeroedDay(time.Now()).AddDate(0, 0, -(days - 1))
	rows, err := workoutrepo.TimedSetsSince(ctx, start)
	if err != nil {
		return WorkoutTimingStats{}, err
	}
	stats := ComputeWorkoutTimingStats(rows)
	stats.Days = days
	stats.Start = start
	return stats, nil
}

func zeroedDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type exerciseAccumulator struct {
	id          uint
	name        string
	setGaps     []time.Duration
	transitions []time.Duration
}

// ComputeWorkoutTimingStats turns raw set rows (ordered by workout, then set
// time) into aggregate timing stats. It is pure so it can be unit tested.
func ComputeWorkoutTimingStats(rows []workoutrepo.TimedSetRow) WorkoutTimingStats {
	var workouts, setGaps, transitions []time.Duration
	perExercise := map[uint]*exerciseAccumulator{}
	excluded := 0

	for _, session := range groupByWorkout(rows) {
		sets := sameDaySets(session)
		if len(sets) < 2 {
			excluded++
			continue
		}
		sort.SliceStable(sets, func(i, j int) bool { return sets[i].SetCreatedAt.Before(sets[j].SetCreatedAt) })
		if medianDuration(consecutiveGaps(sets)) < timingBulkEntryMedianGap {
			excluded++
			continue
		}
		duration := sets[len(sets)-1].SetCreatedAt.Sub(sets[0].SetCreatedAt)
		if duration > timingMaxWorkout {
			excluded++
			continue
		}
		workouts = append(workouts, duration)

		lastByLogged := map[uint]time.Time{}
		for i, s := range sets {
			acc := perExercise[s.ExerciseID]
			if acc == nil {
				acc = &exerciseAccumulator{id: s.ExerciseID, name: s.ExerciseName}
				perExercise[s.ExerciseID] = acc
			}
			if prev, ok := lastByLogged[s.LoggedExerciseID]; ok {
				gap := s.SetCreatedAt.Sub(prev)
				if gap >= timingMinSetGap && gap <= timingMaxSetGap {
					setGaps = append(setGaps, gap)
					acc.setGaps = append(acc.setGaps, gap)
				}
			} else if i > 0 {
				// First set of this exercise: time since the latest set of
				// whatever came before it.
				gap := s.SetCreatedAt.Sub(sets[i-1].SetCreatedAt)
				if gap >= timingMinSetGap && gap <= timingMaxTransitionGap {
					transitions = append(transitions, gap)
					acc.transitions = append(acc.transitions, gap)
				}
			}
			lastByLogged[s.LoggedExerciseID] = s.SetCreatedAt
		}
	}

	exercises := make([]ExerciseTimingStat, 0, len(perExercise))
	for _, acc := range perExercise {
		if len(acc.setGaps) == 0 && len(acc.transitions) == 0 {
			continue
		}
		exercises = append(exercises, ExerciseTimingStat{
			ExerciseID:  acc.id,
			Name:        acc.name,
			SetInterval: summarize(acc.setGaps),
			Transition:  summarize(acc.transitions),
		})
	}
	sort.Slice(exercises, func(i, j int) bool { return exercises[i].Name < exercises[j].Name })

	return WorkoutTimingStats{
		Workout:          summarize(workouts),
		SetInterval:      summarize(setGaps),
		Transition:       summarize(transitions),
		Exercises:        exercises,
		ExcludedWorkouts: excluded,
	}
}

func groupByWorkout(rows []workoutrepo.TimedSetRow) [][]workoutrepo.TimedSetRow {
	var out [][]workoutrepo.TimedSetRow
	index := map[uint]int{}
	for _, r := range rows {
		i, ok := index[r.WorkoutLogID]
		if !ok {
			i = len(out)
			index[r.WorkoutLogID] = i
			out = append(out, nil)
		}
		out[i] = append(out[i], r)
	}
	return out
}

func sameDaySets(session []workoutrepo.TimedSetRow) []workoutrepo.TimedSetRow {
	out := make([]workoutrepo.TimedSetRow, 0, len(session))
	for _, r := range session {
		if r.SetCreatedAt.Before(r.LogDate) || !r.SetCreatedAt.Before(r.LogDate.Add(timingSameDayWindow)) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func consecutiveGaps(sets []workoutrepo.TimedSetRow) []time.Duration {
	gaps := make([]time.Duration, 0, len(sets))
	for i := 1; i < len(sets); i++ {
		gaps = append(gaps, sets[i].SetCreatedAt.Sub(sets[i-1].SetCreatedAt))
	}
	return gaps
}

func medianDuration(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), ds...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func summarize(ds []time.Duration) DurationStat {
	if len(ds) == 0 {
		return DurationStat{}
	}
	var total time.Duration
	for _, d := range ds {
		total += d
	}
	return DurationStat{
		Count:         len(ds),
		AvgSeconds:    (total / time.Duration(len(ds))).Seconds(),
		MedianSeconds: medianDuration(ds).Seconds(),
	}
}
