// Command devseed fills an empty dev database with sample data so every page
// is reachable (workout plans + exercises -> logging, foods, diet days, ...).
//
// It makes one cheap check: if any exercise already exists the DB is treated as
// seeded and nothing is touched, so it is safe to run on every `dev start`.
// Refuses to run when APP_ENV is production.
//
// Usage: go run ./cmd/devseed   (from backend/)
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	authm "be-simpletracker/internal/core/auth/models"
	dietm "be-simpletracker/internal/core/diet/models"
	mediam "be-simpletracker/internal/core/media/models"
	moneym "be-simpletracker/internal/core/money/models"
	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/water"
	"be-simpletracker/internal/core/tracking/weight"
	workm "be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/env"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "devseed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := env.Load(); err != nil {
		return err
	}
	if env.IsProduction() {
		return errors.New("refusing to run with APP_ENV=production")
	}
	db, err := database.ConnectOwner()
	if err != nil {
		return err
	}

	var n int64
	if err := db.Model(&workm.Exercise{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		fmt.Println("devseed: data already present, skipping")
		return nil
	}

	// Owned rows belong to the dev login (devuser creates it; `dev start` runs devuser first).
	owner, err := devOwner(db)
	if err != nil {
		return err
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return seed(tx, owner) }); err != nil {
		return err
	}
	fmt.Println("devseed: sample data created")
	return nil
}

// Local midnight, the same day boundary the API uses (clock.ZerodTime). UTC midnight lands on the
// previous local day west of UTC, so seeded logs would show up a day early.
func today() time.Time {
	n := time.Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.Local)
}

// devOwner is the ID of the dev login (DEV_AUTH_USER, default "dev"), which owns every seeded
// owned row. devseed runs on the owner role with no request identity, so it sets user_id itself
// (database.Owner's create hook accepts a preset owner only when there is no identity).
func devOwner(db *gorm.DB) (uint, error) {
	username := env.StringOr("DEV_AUTH_USER", "dev")
	var user authm.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("user %q not found: run devuser first (dev start does)", username)
		}
		return 0, err
	}
	return user.ID, nil
}

// seed creates the sample data. owner is the dev login: owned tables set user_id to it (lanes
// converting a table pass it to that module's seed function).
func seed(tx *gorm.DB, owner uint) error {
	tracking := func(tx *gorm.DB) error { return seedTracking(tx, owner) }
	media := func(tx *gorm.DB) error { return seedMedia(tx, owner) }
	money := func(tx *gorm.DB) error { return seedMoney(tx, owner) }
	gym := func(tx *gorm.DB) error { return seedGym(tx, owner) }
	diet := func(tx *gorm.DB) error { return seedDiet(tx, owner) }
	for _, f := range []func(*gorm.DB) error{gym, diet, tracking, media, money} {
		if err := f(tx); err != nil {
			return err
		}
	}
	return nil
}

func seedGym(tx *gorm.DB, owner uint) error {
	type ex struct {
		name string
		lt   workm.ExerciseLoadType
	}
	defs := []ex{
		{"Barbell Bench Press", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Overhead Press", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Incline Dumbbell Press", workm.ExerciseLoadTypeFreeWeights},
		{"Triceps Pushdown", workm.ExerciseLoadTypeWeightStack},
		{"Barbell Row", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Lat Pulldown", workm.ExerciseLoadTypeWeightStack},
		{"Dumbbell Curl", workm.ExerciseLoadTypeFreeWeights},
		{"Face Pull", workm.ExerciseLoadTypeWeightStack},
		{"Back Squat", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Romanian Deadlift", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Leg Press", workm.ExerciseLoadTypePlateLoadedWithoutBar},
		{"Leg Curl", workm.ExerciseLoadTypeWeightStack},
	}
	exs := make([]workm.Exercise, len(defs))
	for i, d := range defs {
		exs[i] = workm.Exercise{Creator: database.Creator{CreatedBy: owner}, Name: d.name, LoadType: d.lt, RepRollover: 12, Cues: "Controlled tempo, full range of motion."}
	}
	if err := tx.Create(&exs).Error; err != nil {
		return err
	}

	// The dev login's own machine settings on one shared exercise (exercise_settings).
	if err := tx.Create(&workm.ExerciseSetting{Owner: database.Owner{UserID: owner}, ExerciseID: exs[5].ID,
		MachineSettings: []workm.MachineSetting{{Label: "Seat", Value: "4"}, {Label: "Pad", Value: "2"}}}).Error; err != nil {
		return err
	}

	prog := workm.WorkoutProgram{Owner: database.Owner{UserID: owner}, Name: "Push Pull Legs", IsActive: true}
	if err := tx.Create(&prog).Error; err != nil {
		return err
	}
	// One active program, as ActivateWorkoutProgram keeps it (the baseline seeds an active "Default").
	if err := tx.Model(&workm.WorkoutProgram{}).Where("id != ?", prog.ID).Update("is_active", false).Error; err != nil {
		return err
	}

	type pl struct {
		name    string
		day     int // time.Weekday
		idx     []int
		cardio  string
		minutes int
	}
	plans := []pl{
		{"Push", 1, []int{0, 1, 2, 3}, "", 0},
		{"Pull", 3, []int{4, 5, 6, 7}, "Bike", 15},
		{"Legs", 5, []int{8, 9, 10, 11}, "", 0},
	}
	created := make([]workm.WorkoutPlan, len(plans))
	for i, p := range plans {
		day := p.day
		wp := workm.WorkoutPlan{
			Owner: database.Owner{UserID: owner},
			Name:  p.name, WorkoutProgramID: &prog.ID, DayOfWeek: &day,
			PlannedCardioType: p.cardio, PlannedCardioMinutes: p.minutes,
			PreMobilityItems:  []string{"Arm circles", "Cat-cow"},
			PostMobilityItems: []string{"Hamstring stretch", "Doorway chest stretch"},
		}
		if err := tx.Create(&wp).Error; err != nil {
			return err
		}
		created[i] = wp
		if err := tx.Create(&workm.WorkoutPlanDay{Owner: database.Owner{UserID: owner}, WorkoutPlanID: wp.ID, DayOfWeek: p.day}).Error; err != nil {
			return err
		}
		for order, ei := range p.idx {
			if err := tx.Create(&workm.WorkoutPlanExercise{Owner: database.Owner{UserID: owner}, WorkoutPlanID: wp.ID, ExerciseID: exs[ei].ID, DisplayOrder: order}).Error; err != nil {
				return err
			}
		}
	}

	// Past sessions on each plan's scheduled weekday over the last four weeks (today excluded), with
	// weights rising session by session. One session is skipped so the UI shows a missed day.
	byWeekday := map[time.Weekday]int{}
	for i, p := range plans {
		byWeekday[time.Weekday(p.day)] = i
	}
	const daysBack = 28
	skipped := false
	session := 0
	for back := daysBack; back >= 1; back-- {
		date := today().AddDate(0, 0, -back)
		pi, ok := byWeekday[date.Weekday()]
		if !ok {
			continue
		}
		if !skipped && back > 7 && back <= 14 {
			skipped = true
			continue
		}
		session++
		p := plans[pi]
		wp := created[pi]
		log := workm.WorkoutLog{Owner: database.Owner{UserID: owner}, Date: date, WorkoutPlanID: &wp.ID}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		for _, ei := range p.idx {
			le := workm.LoggedExercise{Owner: database.Owner{UserID: owner}, WorkoutLogID: log.ID, ExerciseID: exs[ei].ID}
			if err := tx.Create(&le).Error; err != nil {
				return err
			}
			base := float32(45 + ei*10 + session*2)
			for s := 0; s < 3; s++ {
				set := workm.LoggedSet{Owner: database.Owner{UserID: owner}, LoggedExerciseID: le.ID, Reps: uint(10 - s), Weight: base, WeightSetup: ""}
				if err := tx.Create(&set).Error; err != nil {
					return err
				}
			}
		}
		if p.cardio != "" {
			c := workm.Cardio{Owner: database.Owner{UserID: owner}, WorkoutLogID: log.ID, Minutes: p.minutes, Type: p.cardio}
			if err := tx.Create(&c).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedDiet(tx *gorm.DB, ownerID uint) error {
	own := database.Owner{UserID: ownerID}
	added := database.Creator{CreatedBy: ownerID}
	type fd struct {
		name, serving          string
		amt, cal, p, f, c, fat float32
	}
	defs := []fd{
		{"Chicken Breast", "g", 100, 165, 31, 0, 0, 3.6},
		{"White Rice (cooked)", "g", 100, 130, 2.7, 0.4, 28, 0.3},
		{"Egg", "each", 1, 72, 6.3, 0, 0.4, 4.8},
		{"Oats", "g", 40, 150, 5, 4, 27, 3},
		{"Banana", "each", 1, 105, 1.3, 3.1, 27, 0.4},
		{"Greek Yogurt", "g", 170, 100, 17, 0, 6, 0.7},
		{"Broccoli", "g", 100, 34, 2.8, 2.6, 7, 0.4},
		{"Olive Oil", "tbsp", 1, 119, 0, 0, 0, 13.5},
		{"Salmon", "g", 100, 208, 20, 0, 0, 13},
		{"Whey Protein", "scoop", 1, 120, 24, 0, 3, 1.5},
	}
	// Item Amount below is a number of servings (ServingAmount units each), not grams.
	foods := make([]dietm.Food, len(defs))
	for i, d := range defs {
		foods[i] = dietm.Food{Creator: added, Name: d.name, ServingType: d.serving, ServingAmount: d.amt, Calories: d.cal, Protein: d.p, Fiber: d.f, Carbs: d.c, Fat: d.fat}
	}
	if err := tx.Create(&foods).Error; err != nil {
		return err
	}

	cf := dietm.CompositeFood{Owner: own, Name: "Protein Oats", Items: []dietm.CompositeFoodItem{
		{Owner: own, FoodID: foods[3].ID, Amount: 1}, {Owner: own, FoodID: foods[9].ID, Amount: 1}, {Owner: own, FoodID: foods[4].ID, Amount: 1},
	}}
	if err := tx.Create(&cf).Error; err != nil {
		return err
	}

	saved := []dietm.SavedMeal{
		{Owner: own, Name: "Chicken & Rice", Items: []dietm.SavedMealItem{{Owner: own, FoodID: foods[0].ID, Amount: 2}, {Owner: own, FoodID: foods[1].ID, Amount: 2.5}, {Owner: own, FoodID: foods[6].ID, Amount: 1}}},
		{Owner: own, Name: "Breakfast", Items: []dietm.SavedMealItem{{Owner: own, FoodID: foods[2].ID, Amount: 3}, {Owner: own, FoodID: foods[4].ID, Amount: 1}}},
	}
	if err := tx.Create(&saved).Error; err != nil {
		return err
	}

	from := today().AddDate(0, 0, -30)
	plan := dietm.Plan{Owner: own, Name: "Maintenance", Calories: 2500, Protein: 180, Fiber: 35, Carbs: 280, Fat: 70, EffectiveFrom: &from}
	if err := tx.Create(&plan).Error; err != nil {
		return err
	}

	// Days from two weeks ago through tomorrow, each with a logged meal; today and
	// tomorrow also get a planned (unlogged) meal.
	for i := -14; i <= 1; i++ {
		day := dietm.DietDay{Owner: own, Date: today().AddDate(0, 0, i), PlanID: plan.ID}
		if err := tx.Create(&day).Error; err != nil {
			return err
		}
		if i <= 0 {
			meal := dietm.Meal{Owner: own, Name: "Lunch", Items: []dietm.MealItem{
				{Owner: own, FoodID: foods[0].ID, Amount: 2}, {Owner: own, FoodID: foods[1].ID, Amount: 2.5}, {Owner: own, FoodID: foods[7].ID, Amount: 1},
			}}
			if err := tx.Create(&meal).Error; err != nil {
				return err
			}
			if err := tx.Create(&dietm.DayLog{Owner: own, DayID: day.ID, MealID: meal.ID}).Error; err != nil {
				return err
			}
		}
		if i >= 0 {
			meal := dietm.Meal{Owner: own, Name: "Dinner", Items: []dietm.MealItem{{Owner: own, FoodID: foods[8].ID, Amount: 1.8}, {Owner: own, FoodID: foods[6].ID, Amount: 1.5}}}
			if err := tx.Create(&meal).Error; err != nil {
				return err
			}
			if err := tx.Create(&dietm.PlannedMeal{Owner: own, DayID: day.ID, MealID: meal.ID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedTracking(tx *gorm.DB, owner uint) error {
	for _, p := range []water.DrinkSizePreset{{Name: "Glass", AmountOz: 8}, {Name: "Bottle", AmountOz: 24}} {
		p.Owner = database.Owner{UserID: owner}
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
	}
	for i := 30; i >= 0; i-- {
		d := today().AddDate(0, 0, -i)
		if err := tx.Create(&weight.BodyWeightLog{Owner: database.Owner{UserID: owner}, Date: d, WeightLbs: 182 - float64(30-i)*0.1 + float64(i%3)*0.4}).Error; err != nil {
			return err
		}
		if err := tx.Create(&steps.StepLog{Owner: database.Owner{UserID: owner}, Date: d, Steps: 6000 + (i*937)%6000}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedMedia adds the shared catalog entries and the dev login's saves of them (both created by owner).
func seedMedia(tx *gorm.DB, owner uint) error {
	year := func(y int) *int { return &y }
	rating := func(r float64) *float64 { return &r }
	type seeded struct {
		item mediam.MediaItem
		save mediam.UserMedia
	}
	rows := []seeded{
		{mediam.MediaItem{Type: mediam.MediaTypeMovie, Title: "Blade Runner 2049", Author: "Denis Villeneuve", ReleaseYear: year(2017), Genres: "Sci-Fi"}, mediam.UserMedia{Status: mediam.MediaStatusCompleted, Rating: rating(4.5), Favorite: true, TimesConsumed: 2}},
		{mediam.MediaItem{Type: mediam.MediaTypeTV, Title: "Severance", Author: "Dan Erickson", ReleaseYear: year(2022), Genres: "Thriller"}, mediam.UserMedia{Status: mediam.MediaStatusInProgress}},
		{mediam.MediaItem{Type: mediam.MediaTypeBook, Title: "Project Hail Mary", Author: "Andy Weir", ReleaseYear: year(2021), Genres: "Sci-Fi"}, mediam.UserMedia{Status: mediam.MediaStatusCompleted, Rating: rating(5)}},
		{mediam.MediaItem{Type: mediam.MediaTypeGame, Title: "Hades", Author: "Supergiant Games", ReleaseYear: year(2020), Genres: "Roguelike"}, mediam.UserMedia{Status: mediam.MediaStatusOnHold}},
		{mediam.MediaItem{Type: mediam.MediaTypeAnime, Title: "Frieren", Author: "Madhouse", ReleaseYear: year(2023), Genres: "Fantasy"}, mediam.UserMedia{Status: mediam.MediaStatusPlanned}},
		{mediam.MediaItem{Type: mediam.MediaTypeAlbum, Title: "In Rainbows", Author: "Radiohead", ReleaseYear: year(2007), Genres: "Rock"}, mediam.UserMedia{Status: mediam.MediaStatusDropped}},
	}
	for _, r := range rows {
		r.item.CreatedBy = owner
		if err := tx.Omit(clause.Associations).Create(&r.item).Error; err != nil {
			return err
		}
		r.save.UserID, r.save.MediaItemID = owner, r.item.ID
		if err := tx.Omit(clause.Associations).Create(&r.save).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedMoney(tx *gorm.DB, owner uint) error {
	own := database.Owner{UserID: owner}
	accts := []moneym.Account{
		{Owner: own, Name: "Everyday Chequing", Kind: moneym.AccountKindChequing, Institution: "Sample Bank", CurrentBalance: 3200},
		{Owner: own, Name: "Savings", Kind: moneym.AccountKindSavings, Institution: "Sample Bank", CurrentBalance: 12000},
		{Owner: own, Name: "Brokerage", Kind: moneym.AccountKindInvestment, Institution: "Sample Brokerage", CurrentBalance: 18500},
	}
	if err := tx.Create(&accts).Error; err != nil {
		return err
	}
	for _, a := range accts {
		for m := 5; m >= 0; m-- {
			b := moneym.AccountBalance{Owner: own, AccountID: a.ID, Date: today().AddDate(0, -m, 0), Balance: a.CurrentBalance - float64(m)*250}
			if err := tx.Create(&b).Error; err != nil {
				return err
			}
		}
	}
	return tx.Create(&moneym.AccountTransaction{Owner: own, AccountID: accts[2].ID, Kind: moneym.TransactionDeposit, Amount: 500, Date: today().AddDate(0, -1, 0), Note: "Monthly contribution"}).Error
}
