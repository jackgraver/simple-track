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
	egg := Food{Name: "Egg", Unit: "Serving", Calories: 140, Protein: 12, Fiber: 0}
	sausage    := Food{Name: "Maple Breakfast Sausage", Unit: "Serving", Calories: 140, Protein: 12, Fiber: 0}
	keto_bread  := Food{Name: "Keto Bread", Unit: "Serving", Calories: 140, Protein: 12, Fiber: 15}
	blueberries := Food{Name: "Blueberries", Unit: "Serving", Calories: 20, Protein: 0.3, Fiber: 0.9}
	kiwi := Food{Name: "Kiwi", Unit: "Piece", Calories: 40, Protein: 0.8, Fiber: 2}

	db.Create(&egg)
	db.Create(&sausage)
	db.Create(&keto_bread)
	db.Create(&blueberries)
	db.Create(&kiwi)

	beef := Food{Name: "Beef", Unit: "Serving", Calories: 200, Protein: 24, Fiber: 0} 
	rice    := Food{Name: "Rice", Unit: "Serving", Calories: 80, Protein: 0, Fiber: 0}
	vegetables  := Food{Name: "Vegetables", Unit: "Serving", Calories: 50, Protein: 1, Fiber: 2}

	db.Create(&beef)
	db.Create(&rice)
	db.Create(&vegetables)
	
	breakfast := Meal{
		Name: "Egg & Sausage Breakfast",
		Items: []MealItem{
			{FoodID: egg.ID, Amount: 1},
			{FoodID: sausage.ID, Amount: 2},
			{FoodID: keto_bread.ID, Amount: 1},
			{FoodID: blueberries.ID, Amount: 1},
			{FoodID: kiwi.ID, Amount: 1},
		},
	}

	dinner := Meal{
		Name: "Ground Beef Bowl",
		Items: []MealItem{
			{FoodID: beef.ID, Amount: 1},
			{FoodID: rice.ID, Amount: 1},
			{FoodID: vegetables.ID, Amount: 1},
		},
	}
	
	db.Create(&breakfast)
	db.Create(&dinner)

	year := 2025
	month := time.September
	daysInMonth := 30 // September has 30 days

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

		mpd := MealPlanDay{
			Date: date,
			Meals: []DayMeal{
				{MealID: breakfast.ID, Status: "expected"},
				{MealID: dinner.ID, Status: "expected"},
			},
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

