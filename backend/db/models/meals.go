package models

type DayMeal struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	DayID  uint   `gorm:"not null" json:"day_id"`
	MealID uint   `gorm:"not null" json:"meal_id"`
	Meal   Meal   `json:"meal"`
	Status string `gorm:"not null;default:'expected'" json:"status"`
	BaseModel
}
type Meal struct {
	ID    uint       `gorm:"primaryKey" json:"id"`
	Name  string     `gorm:"not null" json:"name"`
	Items []MealItem `gorm:"foreignKey:MealID" json:"items"`
	BaseModel
}

// MealItem represents items in a meal (e.g. "200g Beef", "100g Rice")
type MealItem struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	MealID uint    `gorm:"not null" json:"meal_id"`
	FoodID uint    `gorm:"not null" json:"food_id"`
	Food   Food    `gorm:"foreignKey:FoodID" json:"food"`
	Amount float64 `gorm:"not null" json:"amount"` // e.g. grams
	BaseModel
}

type Food struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	Name     string  `gorm:"not null" json:"name"`
	Unit     string  `gorm:"not null" json:"unit"`
	Calories float32 `gorm:"not null" json:"calories"` // per unit
	Protein  float32 `json:"protein"`
	Fiber    float32 `json:"fiber"`
	BaseModel
}