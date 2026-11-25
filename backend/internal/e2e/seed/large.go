package seed

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	dietm "be-simpletracker/internal/core/diet/models"
	mediam "be-simpletracker/internal/core/media/models"
	moneym "be-simpletracker/internal/core/money/models"
	"be-simpletracker/internal/core/tracking/grocery"
	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/weight"
	workm "be-simpletracker/internal/core/workout/models"
)

// Profiles are extra data layered on the base seed, for scale and edge-case tests (TASKS C3).
// Apply one after Reset (POST /__e2e/seed/<profile>, or `e2eseed -profile <name>`).
var Profiles = map[string]func(*gorm.DB) error{
	"large": Large,
}

// LargeDays is how far back the large profile's history goes from EpochDay.
const LargeDays = 730

// Large adds two years of history in every module, plus awkward values the UI must survive:
// unicode and very long names, fractional weights, zero and very large numbers.
// Deterministic: no randomness, everything derived from the day index.
func Large(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, f := range []func(*gorm.DB) error{largeGym, largeDiet, largeTracking, largeMedia, largeMoney, largeGrocery} {
			if err := f(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

func day(back int) time.Time { return EpochDay().AddDate(0, 0, -back) }

const batch = 500

func largeGym(tx *gorm.DB) error {
	var plans []workm.WorkoutPlan
	if err := tx.Order("id").Find(&plans).Error; err != nil {
		return err
	}
	var pes []workm.WorkoutPlanExercise
	if err := tx.Order("workout_plan_id, display_order").Find(&pes).Error; err != nil {
		return err
	}
	byPlan := map[uint][]uint{}
	for _, pe := range pes {
		byPlan[pe.WorkoutPlanID] = append(byPlan[pe.WorkoutPlanID], pe.ExerciseID)
	}
	byWeekday := map[time.Weekday]workm.WorkoutPlan{}
	for _, p := range plans {
		if p.DayOfWeek != nil {
			byWeekday[time.Weekday(*p.DayOfWeek)] = p
		}
	}

	// Every scheduled day before today, except every 9th session (missed days).
	var logs []workm.WorkoutLog
	session := 0
	for back := LargeDays; back >= 1; back-- {
		p, ok := byWeekday[day(back).Weekday()]
		if !ok {
			continue
		}
		session++
		if session%9 == 0 {
			continue
		}
		id := p.ID
		logs = append(logs, workm.WorkoutLog{Date: day(back), WorkoutPlanID: &id})
	}
	if err := tx.CreateInBatches(&logs, batch).Error; err != nil {
		return fmt.Errorf("workout logs: %w", err)
	}

	var les []workm.LoggedExercise
	for _, l := range logs {
		for _, ex := range byPlan[*l.WorkoutPlanID] {
			les = append(les, workm.LoggedExercise{WorkoutLogID: l.ID, ExerciseID: ex})
		}
	}
	if err := tx.CreateInBatches(&les, batch).Error; err != nil {
		return fmt.Errorf("logged exercises: %w", err)
	}

	sets := make([]workm.LoggedSet, 0, len(les)*3)
	for i, le := range les {
		// Slow progression in 2.5 lb steps, with a deload every ~3 months.
		base := float32(45+int(le.ExerciseID)*10) + 2.5*float32((i/len(byWeekday))%40)
		for s := 0; s < 3; s++ {
			sets = append(sets, workm.LoggedSet{LoggedExerciseID: le.ID, Reps: uint(10 - s), Weight: base})
		}
	}
	// Edge values on the oldest session: a zero-weight set and a very heavy one.
	if len(sets) >= 2 {
		sets[0].Weight, sets[0].Reps = 0, 25
		sets[1].Weight, sets[1].Reps = 1000, 1
	}
	if err := tx.CreateInBatches(&sets, batch).Error; err != nil {
		return fmt.Errorf("logged sets: %w", err)
	}
	return nil
}

func largeDiet(tx *gorm.DB) error {
	odd := []dietm.Food{
		{Name: "Crème fraîche 🥛", ServingType: "g", ServingAmount: 30, Calories: 101, Protein: 0.6, Fat: 10.8},
		{Name: "Extra-long food name that keeps going to check wrapping and truncation in every list", ServingType: "each", ServingAmount: 1, Calories: 250, Protein: 12},
		{Name: "Water (zero calories)", ServingType: "ml", ServingAmount: 250, Calories: 0},
		{Name: "Bulk shake", ServingType: "scoop", ServingAmount: 1, Calories: 1850, Protein: 110, Carbs: 250, Fat: 40},
	}
	if err := tx.Create(&odd).Error; err != nil {
		return err
	}
	var foods []dietm.Food
	if err := tx.Order("id").Limit(8).Find(&foods).Error; err != nil {
		return err
	}
	var plan dietm.Plan
	if err := tx.First(&plan).Error; err != nil {
		return err
	}

	days := make([]dietm.DietDay, 0, LargeDays)
	for back := LargeDays; back >= 1; back-- {
		days = append(days, dietm.DietDay{Date: day(back), PlanID: plan.ID})
	}
	if err := tx.CreateInBatches(&days, batch).Error; err != nil {
		return fmt.Errorf("diet days: %w", err)
	}
	meals := make([]dietm.Meal, 0, len(days))
	for i := range days {
		a, b := foods[i%len(foods)], foods[(i+3)%len(foods)]
		meals = append(meals, dietm.Meal{Name: "Lunch", Items: []dietm.MealItem{
			{FoodID: a.ID, Amount: 1 + float32(i%3)}, {FoodID: b.ID, Amount: 1.5},
		}})
	}
	if err := tx.CreateInBatches(&meals, batch).Error; err != nil {
		return fmt.Errorf("meals: %w", err)
	}
	logs := make([]dietm.DayLog, len(days))
	for i := range days {
		logs[i] = dietm.DayLog{DayID: days[i].ID, MealID: meals[i].ID}
	}
	return tx.CreateInBatches(&logs, batch).Error
}

func largeTracking(tx *gorm.DB) error {
	ws := make([]weight.BodyWeightLog, 0, LargeDays)
	ss := make([]steps.StepLog, 0, LargeDays)
	for back := LargeDays; back >= 0; back-- {
		// A slow cut, a bulk, and day-to-day noise; deterministic.
		w := 190 - 12*math.Sin(float64(back)/120) + float64(back%5)*0.3
		if back%11 != 0 { // gaps: not every day is weighed
			ws = append(ws, weight.BodyWeightLog{Date: day(back), WeightLbs: math.Round(w*10) / 10})
		}
		n := 2000 + (back*7919)%14000
		if back%97 == 0 {
			n = 0
		}
		if back%101 == 0 {
			n = 48000
		}
		ss = append(ss, steps.StepLog{Date: day(back), Steps: n})
	}
	if err := tx.CreateInBatches(&ws, batch).Error; err != nil {
		return err
	}
	return tx.CreateInBatches(&ss, batch).Error
}

func largeMedia(tx *gorm.DB) error {
	types := []mediam.MediaType{mediam.MediaTypeMovie, mediam.MediaTypeTV, mediam.MediaTypeBook, mediam.MediaTypeGame, mediam.MediaTypeAnime, mediam.MediaTypeAlbum}
	statuses := []mediam.MediaStatus{mediam.MediaStatusPlanned, mediam.MediaStatusInProgress, mediam.MediaStatusCompleted, mediam.MediaStatusOnHold, mediam.MediaStatusDropped}
	items := make([]mediam.MediaItem, 0, 300)
	for i := 0; i < 300; i++ {
		year := 1950 + i%75
		it := mediam.MediaItem{
			Type: types[i%len(types)], Title: fmt.Sprintf("Title %03d", i), Creator: fmt.Sprintf("Creator %d", i%40),
			ReleaseYear: &year, Genres: "Drama", Status: statuses[i%len(statuses)], Favorite: i%17 == 0,
		}
		if it.Status == mediam.MediaStatusCompleted {
			r := float64(1+i%10) / 2
			it.Rating = &r
		}
		items = append(items, it)
	}
	items[0].Title = "Amélie: Le Fabuleux Destin d’Amélie Poulain 🎬"
	items[1].Title = "A very long title that goes on and on to check how cards and rows cope with overflow"
	return tx.CreateInBatches(&items, batch).Error
}

func largeMoney(tx *gorm.DB) error {
	accts := []moneym.Account{
		{Name: "Everyday Chequing", Kind: moneym.AccountKindChequing, Institution: "Sample Bank", CurrentBalance: 3200},
		{Name: "Savings", Kind: moneym.AccountKindSavings, Institution: "Sample Bank", CurrentBalance: 12000},
		{Name: "Brokerage", Kind: moneym.AccountKindInvestment, Institution: "Sample Brokerage", CurrentBalance: 1_250_000.55},
		{Name: "Overdrawn", Kind: moneym.AccountKindChequing, Institution: "Sample Bank", CurrentBalance: -420.10},
	}
	if err := tx.Create(&accts).Error; err != nil {
		return err
	}
	var bals []moneym.AccountBalance
	var txs []moneym.AccountTransaction
	for _, a := range accts {
		for m := 24; m >= 0; m-- {
			bals = append(bals, moneym.AccountBalance{AccountID: a.ID, Date: EpochDay().AddDate(0, -m, 0), Balance: a.CurrentBalance - float64(m)*137.5})
		}
	}
	for m := 24; m >= 1; m-- {
		txs = append(txs, moneym.AccountTransaction{AccountID: accts[2].ID, Kind: moneym.TransactionDeposit, Amount: 500, Date: EpochDay().AddDate(0, -m, 0), Note: "Monthly contribution"})
	}
	if err := tx.CreateInBatches(&bals, batch).Error; err != nil {
		return err
	}
	return tx.CreateInBatches(&txs, batch).Error
}

func largeGrocery(tx *gorm.DB) error {
	items := make([]grocery.GroceryItem, 0, 150)
	for i := 0; i < 150; i++ {
		it := grocery.GroceryItem{Name: fmt.Sprintf("Item %03d", i)}
		if i%3 != 0 { // most are history (completed), which feeds suggestions
			t := day(i % 60)
			it.CompletedAt = &t
		}
		items = append(items, it)
	}
	items[0].Name = "Jalapeño 🌶️ (fresh, not pickled, the big ones from the corner store)"
	return tx.CreateInBatches(&items, batch).Error
}
