package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MealPlanModel struct{
	db *gorm.DB
}

func NewMealPlanModel(db *gorm.DB) *MealPlanModel {
	return &MealPlanModel{db: db}
}

func (m *MealPlanModel) MigrateDatabase() {
	fmt.Println("Migrating meal plan database")
	if err := m.db.Migrator().DropTable(
		&DayLog{},
		&PlannedMeal{},
		&MealItem{},
		&SavedMealItem{},
		&Meal{},
		&SavedMeal{},
		&Food{},
		&Day{},
		&Plan{},
	); err != nil {
		fmt.Printf("Failed to drop meal plan database: %v\n", err)
	}

	if err := m.db.AutoMigrate(
		&Plan{},
		&Day{},
		&Meal{},
		&MealItem{},
		&SavedMeal{},
		&SavedMealItem{},
		&PlannedMeal{},
		&DayLog{},
		&Food{},
	); err != nil {
		fmt.Printf("Failed to migrate meal plan database: %v\n", err)
	}

	if err := m.seedDatabase(m.db); err != nil {
		fmt.Printf("Failed to seed meal plan database: %v\n", err)
	}
}

func (m *MealPlanModel) seedDatabase(db *gorm.DB) error {
	fmt.Println("Seeding meal plan database")

	// egg := Food{Name: "Egg", ServingType: "unit", ServingAmount: 2, Calories: 160, Protein: 12, Fiber: 0, Carbs: 0}
	// sausage    := Food{Name: "Maple Breakfast Sausage", ServingType: "unit", ServingAmount: 3, Calories: 140, Protein: 12, Fiber: 0, Carbs: 0}
	// keto_bread  := Food{Name: "Keto Bread", Calories: 140, ServingType: "unit", ServingAmount: 2, Protein: 12, Fiber: 15, Carbs: 0}
	// blueberries := Food{Name: "Blueberries", ServingType: "g", ServingAmount: 50, Calories: 20, Protein: 0.3, Fiber: 0.9, Carbs: 0}
	// kiwi := Food{Name: "Kiwi", Calories: 40, ServingType: "unit", ServingAmount: 1, Protein: 0.8, Fiber: 2, Carbs: 0}

	// foods := []*Food{&egg, &sausage, &keto_bread, &blueberries, &kiwi}
	// db.Create(&foods)

	// beef := Food{Name: "Beef", ServingType: "g", Calories: 200, Protein: 24, Fiber: 0, Carbs: 0} 
	// rice    := Food{Name: "Rice", ServingType: "g", Calories: 80, Protein: 0, Fiber: 0, Carbs: 0}
	// vegetables  := Food{Name: "Vegetables", ServingType: "g", Calories: 50, Protein: 1, Fiber: 2, Carbs: 0}

	// db.Create(&beef)
	// db.Create(&rice)
	// db.Create(&vegetables)
	
	// breakfast := Meal{
	// 	Name: "Egg & Sausage Breakfast",
	// 	Items: []MealItem{
	// 		{FoodID: egg.ID, Amount: 2},
	// 		{FoodID: sausage.ID, Amount: 6},
	// 		{FoodID: keto_bread.ID, Amount: 2},
	// 		{FoodID: blueberries.ID, Amount: 40},
	// 		{FoodID: kiwi.ID, Amount: 1},
	// 	},
	// }

	// dinner := Meal{
	// 	Name: "Ground Beef Bowl",
	// 	Items: []MealItem{
	// 		{FoodID: beef.ID, Amount: 1},
	// 		{FoodID: rice.ID, Amount: 1},
	// 		{FoodID: vegetables.ID, Amount: 1},
	// 	},
	// }
	
	// m2 := Meal{
	// 	Name: "Meal 2",
	// 	Items: []MealItem{
	// 		{FoodID: beef.ID, Amount: 1},
	// 		{FoodID: keto_bread.ID, Amount: 1},
	// 		{FoodID: blueberries.ID, Amount: 1},
	// 	},
	// }

	// m3 := Meal{
	// 	Name: "Meal 3",
	// 	Items: []MealItem{
	// 		{FoodID: egg.ID, Amount: 1},
	// 		{FoodID: rice.ID, Amount: 1},
	// 		{FoodID: vegetables.ID, Amount: 1},
	// 		{FoodID: blueberries.ID, Amount: 1},
	// 		{FoodID: kiwi.ID, Amount: 1},
	// 	},
	// }
	
	// db.Create(&breakfast)
	// db.Create(&dinner)
	// db.Create(&m2)
	// db.Create(&m3)

	// cut := Plan{Name: "Cut",
	// 			Calories: 1400,
	// 			Protein:  150,
	// 			Fiber:    30,
	// 			Carbs:    150,
	// 		}
	// db.Create(&cut)
	bulk := Plan{Name: "Bulk",
				Calories: 2400,
				Protein:  150,
				Fiber:    50,
				Carbs:    150,
			}
	db.Create(&bulk)

	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.Local)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		mpd := Day{
			Date: date,
			Plan: bulk,
			PlannedMeals: []PlannedMeal{},
		}

		if err := db.Create(&mpd).Error; err != nil {
			fmt.Printf("Failed to create day %v: %v\n", date, err)
		}
	}
	return nil;
}

func (m *MealPlanModel) Preloads() []string {
	return []string{"PlannedMeals.Meal.Items.Food", "Plan", "Logs.Meal.Items.Food"}
}

type Day struct {
    gorm.Model
    Date   time.Time `json:"date"`
    PlanID uint `json:"plan_id"`            // FK to Plan
    Plan   Plan `gorm:"foreignKey:PlanID" json:"plan"`   // the Plan object
    PlannedMeals []PlannedMeal `json:"plannedMeals"`
    Logs         []DayLog `json:"loggedMeals"`
}

func (d *Day) Preloads() []string {
    return []string{"PlannedMeals.Meal.Items.Food", "Plan", "Logs.Meal.Items.Food"}
}

type Plan struct {
    gorm.Model
    Name string `json:"name"`
	Calories float32 `json:"calories"`
    Protein  float32 `json:"protein"`
    Fiber    float32 `json:"fiber"`
	Carbs    float32 `json:"carbs"`
}

type PlannedMeal struct {
    gorm.Model
    DayID  uint `json:"day_id" gorm:"not null"`
    Day    Day  `json:"day"`
    MealID uint `json:"meal_id" gorm:"not null"`
    Meal   Meal `json:"meal"`
	Logged bool `json:"logged"`
}

type DayLog struct {
    gorm.Model
    DayID uint `json:"day_id" gorm:"not null"`
    MealID uint `json:"meal_id" gorm:"not null"`
    Meal   Meal `json:"meal"`
}

type Meal struct {
    gorm.Model
    Name  string     `json:"name" gorm:"not null"`
    Items []MealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

type MealItem struct {
    gorm.Model
    MealID uint `json:"meal_id" gorm:"not null;index"`
    FoodID uint `json:"food_id" gorm:"not null;index"`
    Amount float32 `json:"amount"`

    Meal Meal `json:"meal"`
    Food Food `json:"food"`
}

type SavedMeal struct {
	gorm.Model
	Name  string     `json:"name" gorm:"not null"`
	Items []SavedMealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

type SavedMealItem struct {
    gorm.Model
    SavedMealID uint `json:"saved_meal_id" gorm:"not null;index"`
    FoodID      uint `json:"food_id" gorm:"not null;index"`
    Amount      float64 `json:"amount"`

    SavedMeal SavedMeal
    Food      Food
}

type Food struct {
    gorm.Model
    Name     string  `json:"name" gorm:"not null;uniqueIndex"`
	ServingType string  `json:"serving_type" gorm:"not null"`
	ServingAmount float32 `json:"serving_amount" gorm:"not null"`
    Calories float32 `json:"calories" gorm:"not null"`
    Protein  float32 `json:"protein"`
    Fiber    float32 `json:"fiber"`
	Carbs    float32 `json:"carbs"`
}