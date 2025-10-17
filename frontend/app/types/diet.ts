export interface BaseModel {
    ID: number;
    created_at: string;
    updated_at: string;
}

export interface Day extends BaseModel {
    date: string;
    plan: Plan;
    plannedMeals: PlannedMeal[];
    loggedMeals: DayLog[];
}

export interface Plan extends BaseModel {
    name: string;
    calories: number;
    protein: number;
    fiber: number;
}

export interface PlannedMeal extends BaseModel {
    day_id: number;
    meal_id: number;
    meal: Meal;
}

export interface DayLog extends BaseModel {
    day_id: number;
    meal_id: number;
    meal: Meal;
}

export interface Meal extends BaseModel {
    name: string;
    items: MealItem[];
}

export interface MealItem extends BaseModel {
    meal_id: number;
    food_id: number;
    food?: Food;
    amount: number;
}

export interface Food extends BaseModel {
    name: string;
    unit: string;
    calories: number;
    protein: number;
    fiber: number;
}

export interface SavedMeal extends BaseModel {
    name: string;
    items: SavedMealItem[];
}

export interface SavedMealItem extends BaseModel {
    saved_meal_id: number;
    saved_meal: SavedMeal;
    saved_food_id: number;
    saved_food: SavedFood;
    amount: number;
}

export interface SavedFood extends BaseModel {
    name: string;
    unit: string;
    calories: number;
    protein: number;
    fiber: number;
}
