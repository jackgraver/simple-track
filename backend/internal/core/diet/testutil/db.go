package testutil

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/database"
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// UserID is the account every seed row here belongs to, and the one Ctx acts as.
const UserID uint = 1

var owner = database.Owner{UserID: UserID}

// Ctx is a request-shaped context for the test user on db (SQLite module tests).
func Ctx(db *gorm.DB) context.Context {
	return database.WithIdentity(database.WithDB(context.Background(), db), database.Identity{UserID: UserID})
}

// CtxAs is Ctx for another user.
func CtxAs(db *gorm.DB, uid uint) context.Context {
	return database.WithIdentity(database.WithDB(context.Background(), db), database.Identity{UserID: uid})
}

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Plan{},
		&models.DietDay{},
		&models.Meal{},
		&models.MealItem{},
		&models.SavedMeal{},
		&models.SavedMealItem{},
		&models.PlannedMeal{},
		&models.DayLog{},
		&models.Food{},
		&models.CompositeFood{},
		&models.CompositeFoodItem{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func SeedPlan(t *testing.T, db *gorm.DB, name string) models.Plan {
	t.Helper()
	p := models.Plan{
		Owner:    owner,
		Name:     name,
		Calories: 2000,
		Protein:  150,
		Fiber:    30,
		Carbs:    200,
		Fat:      65,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	return p
}

func SeedFood(t *testing.T, db *gorm.DB, name string, macros models.Food) models.Food {
	t.Helper()
	f := macros
	f.CreatedBy = UserID
	f.Name = name
	f.ServingType = "g"
	f.ServingAmount = 100
	if err := db.Create(&f).Error; err != nil {
		t.Fatal(err)
	}
	return f
}

func SeedDay(t *testing.T, db *gorm.DB, date time.Time, planID uint) models.DietDay {
	t.Helper()
	d := models.DietDay{
		Owner:  owner,
		Date:   date,
		PlanID: planID,
	}
	if err := db.Create(&d).Error; err != nil {
		t.Fatal(err)
	}
	return d
}

func SeedMeal(t *testing.T, db *gorm.DB, name string, foodID uint, amount float32) models.Meal {
	t.Helper()
	m := models.Meal{
		Owner: owner,
		Name:  name,
		Items: []models.MealItem{{
			Owner:  owner,
			FoodID: foodID,
			Amount: amount,
		}},
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatal(err)
	}
	return m
}

func SeedSavedMeal(t *testing.T, db *gorm.DB, name string, foodID uint) models.SavedMeal {
	t.Helper()
	sm := models.SavedMeal{
		Owner: owner,
		Name:  name,
		Items: []models.SavedMealItem{{
			Owner:  owner,
			FoodID: foodID,
			Amount: 1,
		}},
	}
	if err := db.Create(&sm).Error; err != nil {
		t.Fatal(err)
	}
	return sm
}

func SeedDayLog(t *testing.T, db *gorm.DB, dayID, mealID uint) models.DayLog {
	t.Helper()
	dl := models.DayLog{Owner: owner, DayID: dayID, MealID: mealID}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatal(err)
	}
	return dl
}

func SeedPlannedMeal(t *testing.T, db *gorm.DB, dayID, mealID uint, order int, logged bool) models.PlannedMeal {
	t.Helper()
	pm := models.PlannedMeal{
		Owner:        owner,
		DayID:        dayID,
		MealID:       mealID,
		Logged:       logged,
		DisplayOrder: order,
	}
	if err := db.Create(&pm).Error; err != nil {
		t.Fatal(err)
	}
	return pm
}

func Today() time.Time {
	return clock.ZerodTime(0)
}

func DefaultMacros() models.Food {
	return models.Food{
		Calories: 100,
		Protein:  10,
		Fiber:    2,
		Carbs:    12,
		Fat:      3,
	}
}

// SeedPlannedFromSaved is an unlogged planned meal that came from a saved meal.
func SeedPlannedFromSaved(t *testing.T, db *gorm.DB, dayID, mealID, savedMealID uint) models.PlannedMeal {
	t.Helper()
	pm := models.PlannedMeal{Owner: owner, DayID: dayID, MealID: mealID, SavedMealID: &savedMealID}
	if err := db.Create(&pm).Error; err != nil {
		t.Fatal(err)
	}
	return pm
}
