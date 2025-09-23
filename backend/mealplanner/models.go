package mealplanner

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func MigrateMealPlanDatabase(db *gorm.DB) {
	db.Migrator().DropTable(&Food{}, &Meal{}, &MealItem{}, &DayGoals{}, &MealPlanDay{}, &DayMeal{})
	db.AutoMigrate(&Food{}, &Meal{}, &MealItem{}, &DayGoals{}, &MealPlanDay{}, &DayMeal{})
	seed(db)
}

func seed(db *gorm.DB) {
	chicken := Food{Name: "Chicken Breast", Unit: "g", Calories: 1.65, Protein: 0.31, Fiber: 0}
	rice    := Food{Name: "White Rice", Unit: "g", Calories: 1.30, Protein: 0.02, Fiber: 0.01}
	yogurt  := Food{Name: "Greek Yogurt", Unit: "g", Calories: 0.59, Protein: 0.10, Fiber: 0}
	berries := Food{Name: "Blueberries", Unit: "g", Calories: 0.57, Protein: 0.01, Fiber: 0.025}

	db.Create(&chicken)
	db.Create(&rice)
	db.Create(&yogurt)
	db.Create(&berries)
	
	chickenRice := Meal{
		Name: "Chicken & Rice Bowl",
		Items: []MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	proteinBreakfast := Meal{
		Name: "Protein Breakfast",
		Items: []MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	greekYogurt := Meal{
		Name: "Greek Yogurt Bowl",
		Items: []MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	
	db.Create(&chickenRice)
	db.Create(&proteinBreakfast)
	db.Create(&greekYogurt)

	// 2. Create day with meals and items
	day1 := MealPlanDay{
		Date: time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC),
		Meals: []DayMeal{
			{MealID: chickenRice.ID, Status: "expected"},
			{MealID: chickenRice.ID, Status: "actual"},
		},
		Goals: DayGoals{
			Calories: 2000,
			Protein:  150,
			Fiber:    65,
		},
	}
	db.Create(&day1)
	
	day2 := MealPlanDay{
		Date: time.Date(2025, time.September, 21, 0, 0, 0, 0, time.UTC),
		Meals: []DayMeal{
			{MealID: proteinBreakfast.ID, Status: "expected"},
			{MealID: proteinBreakfast.ID, Status: "actual"},
			{MealID: greekYogurt.ID, Status: "expected"},
		},
		Goals: DayGoals{
			Calories: 1900,
			Protein:  160,
			Fiber:    45,
		},
	}
	db.Create(&day2)

	year := 2025
	month := time.September
	daysInMonth := 30 // September has 30 days

	for day := 1; day <= daysInMonth; day++ {
		if day == 20 || day == 21 {
			continue
		}

		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

		mpd := MealPlanDay{
			Date: date,
			Goals: DayGoals{
				Calories: 2000,
				Protein:  150,
				Fiber:    50,
			},
		}

		if err := db.Create(&mpd).Error; err != nil {
			fmt.Printf("Failed to create day %v: %v\n", date, err)
		}
	}
}

type MealPlanDay struct {
    gorm.Model
    Date  time.Time   `json:"date"`
    Meals []DayMeal   `json:"meals"`
    Goals DayGoals    `json:"goals" gorm:"foreignKey:MealPlanDayID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type DayGoals struct {
    gorm.Model
    MealPlanDayID uint    `json:"meal_plan_day_id" gorm:"uniqueIndex"` // each day has one goals record
    Calories      float32 `json:"calories"`
    Protein       float32 `json:"protein"`
    Fiber         float32 `json:"fiber"`
}

type DayMeal struct {
    gorm.Model
    MealPlanDayID uint   `json:"meal_plan_day_id" gorm:"not null"` // FK back to MealPlanDay
    Meal          Meal   `json:"meal"`
    MealID        uint   `json:"meal_id" gorm:"not null"`          // explicit FK to Meal
    Status        string `json:"status" gorm:"not null;default:'expected'"`
}

type Meal struct {
    gorm.Model
    Name  string     `json:"name" gorm:"not null"`
    Items []MealItem `json:"items"`
}

type MealItem struct {
    gorm.Model
    MealID uint    `json:"meal_id" gorm:"not null"` // belongs to Meal
    FoodID uint    `json:"food_id" gorm:"not null"` // belongs to Food
    Food   Food    `json:"food"`
    Amount float64 `json:"amount"`
}

type Food struct {
    gorm.Model
    Name     string  `json:"name" gorm:"not null"`
    Unit     string  `json:"unit" gorm:"not null"`
    Calories float32 `json:"calories" gorm:"not null"`
    Protein  float32 `json:"protein"`
    Fiber    float32 `json:"fiber"`
}

