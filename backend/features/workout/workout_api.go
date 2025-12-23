package workout

import (
	"be-simpletracker/features/workout/models"
	"be-simpletracker/features/workout/routes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers all workout feature routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/workout")
	
	seedDatabase(h.db)

	// Register plan sub-domain routes
	routes.RegisterWorkoutPlanRoutes(group, h.db)
	routes.RegisterExercisesRoutes(group, h.db)	
	routes.RegisterWorkoutLogRoutes(group, h.db)
}


func seedDatabase(db *gorm.DB) error {
	fmt.Println("Migrating workout database")
    if err := db.Migrator().DropTable(
        &models.WorkoutPlan{},
        &models.WorkoutLog{},
        &models.LoggedExercise{},
		&models.Exercise{},
        &models.LoggedSet{},
        &models.Cardio{},
    ); err != nil {
		fmt.Printf("Failed to drop workout database: %v\n", err)
	}

    if err := db.AutoMigrate(
        &models.WorkoutPlan{},
        &models.WorkoutLog{},
        &models.LoggedExercise{},
		&models.Exercise{},
        &models.LoggedSet{},
        &models.Cardio{},
    ); err != nil {
		fmt.Printf("Failed to migrate workout database: %v\n", err)
	}

	fmt.Println("Seeding workout database")

	inclinePress := models.Exercise{Name: "Incline Press", RepRollover: 10}
	chestFly := models.Exercise{Name: "Chest Fly", RepRollover: 10}
	dips := models.Exercise{Name: "Dips", RepRollover: 10}
	latRaise := models.Exercise{Name: "Lat Raise", RepRollover: 10}
	shoulderPress := models.Exercise{Name: "Shoulder Press", RepRollover: 10}
	squat := models.Exercise{Name: "Squat", RepRollover: 10}
	deadlift := models.Exercise{Name: "Deadlift", RepRollover: 10}
	benchPress := models.Exercise{Name: "Bench Press", RepRollover: 10}
	cableRows := models.Exercise{Name: "Cable Rows", RepRollover: 10}
	barbellRows := models.Exercise{Name: "Barbell Rows", RepRollover: 10}
	facePulls := models.Exercise{Name: "Face Pulls", RepRollover: 10}
	pulldowns := models.Exercise{Name: "Pulldowns", RepRollover: 10}
	JMPress := models.Exercise{Name: "JM Press", RepRollover: 10}
	extensions := models.Exercise{Name: "Extensions", RepRollover: 10}
	inclineCurls := models.Exercise{Name: "Incline Curls", RepRollover: 10}
	hammerCurls := models.Exercise{Name: "Hammer Curls", RepRollover: 10}
	calfRaise := models.Exercise{Name: "Calf Raises", RepRollover: 10}
	abCrunches := models.Exercise{Name: "Ab Crunches", RepRollover: 10}
	legPress := models.Exercise{Name: "Leg Press", RepRollover: 10}
	legExtensions := models.Exercise{Name: "Leg Extensions", RepRollover: 10}
	hamstringCurls := models.Exercise{Name: "Hamstring Curls", RepRollover: 10}
	hipPress := models.Exercise{Name: "Hip Press", RepRollover: 10}
	hipExtensions := models.Exercise{Name: "Hip Extensions", RepRollover: 10}
	outerThigh := models.Exercise{Name: "Outer Thigh", RepRollover: 10}
	innerThigh := models.Exercise{Name: "Inner Thigh", RepRollover: 10}
	hackSquat := models.Exercise{Name: "Hack Squat", RepRollover: 10}

	exercises := []*models.Exercise{
		&inclinePress,
		&chestFly,
		&dips,
		&latRaise,
		&shoulderPress,
		&squat,
		&deadlift,
		&benchPress,
		&cableRows,
		&barbellRows,
		&facePulls,
		&pulldowns,
		&JMPress,
		&extensions,
		&inclineCurls,
		&hammerCurls,
		&calfRaise,
		&abCrunches,
		&legPress,
		&legExtensions,
		&hamstringCurls,
		&hipPress,
		&hipExtensions,
		&outerThigh,
		&innerThigh,
	}

	for _, exercise := range exercises {
		db.Create(exercise)
	}

	push_plan := models.WorkoutPlan{
		Name: "Push",
		Exercises: []models.Exercise{
			inclinePress,
			chestFly,
			dips,
			latRaise,
			shoulderPress,
			JMPress,
			extensions,
		},
	}
	pull_plan := models.WorkoutPlan{
		Name: "Pull",
		Exercises: []models.Exercise{
			barbellRows,
			facePulls,
			pulldowns,
			cableRows,
			inclineCurls,
			hammerCurls,
		},
	}
	legs_plan := models.WorkoutPlan{
		Name: "Legs",
		Exercises: []models.Exercise{
			outerThigh,
			innerThigh,
			legExtensions,
			hamstringCurls,
			squat,
			deadlift,
			calfRaise,
		},
	}
	upper_plan := models.WorkoutPlan{
		Name: "Upper",
		Exercises: []models.Exercise{
			inclinePress,
			chestFly,
			dips,
			latRaise,
			shoulderPress,
			barbellRows,
			pulldowns,
			JMPress,
			extensions,
			inclineCurls,	
			hammerCurls,
		},
	}
	lower_plan := models.WorkoutPlan{
		Name: "Lower",
		Exercises: []models.Exercise{
			outerThigh,
			innerThigh,
			legExtensions,
			hamstringCurls,
			squat,
			deadlift,
			hackSquat,
			calfRaise,
		},
	}
	rest_plan := models.WorkoutPlan{
		Name: "Rest",
	}
	
	db.Create(&push_plan)
	db.Create(&pull_plan)
	db.Create(&legs_plan)
	db.Create(&upper_plan)
	db.Create(&lower_plan)
	db.Create(&rest_plan)

	now := time.Now()
	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location())

	weekdayPlans := map[time.Weekday]models.WorkoutPlan{
		time.Sunday:    lower_plan,
		time.Monday:    rest_plan,
		time.Tuesday:   push_plan,
		time.Wednesday: pull_plan,
		time.Thursday:  legs_plan,
		time.Friday:    rest_plan,
		time.Saturday:  upper_plan,	
	}

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		plan := weekdayPlans[date.Weekday()]
		wl := models.WorkoutLog{
			Date: date,
			WorkoutPlan: &plan,
		}
		db.Create(&wl)
	}
	return nil;
}
