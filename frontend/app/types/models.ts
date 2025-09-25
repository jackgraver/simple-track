export interface BaseModel {
    created_at: string;
    updated_at: string;
}

export interface Food extends BaseModel {
    ID: number;
    name: string;
    unit: string;
    calories: number;
    protein: number;
    fiber: number;
}

export interface MealItem extends BaseModel {
    ID: number;
    meal_id: number;
    food_id: number;
    food?: Food;
    amount: number;
}

export interface Meal extends BaseModel {
    ID: number;
    name: string;
    items: MealItem[];
}

export interface DayMeal extends BaseModel {
    ID: number;
    meal_plan_day_id: number;
    meal_id: number;
    meal: Meal;
    status: string;
}

export interface DayGoals extends BaseModel {
    ID: number;
    day_id: number;
    calories: number;
    protein: number;
    fiber: number;
}

export interface MealPlanDay extends BaseModel {
    ID: number;
    date: string;
    meals: DayMeal[];
    goals: DayGoals;
}

export interface MealLog {
    ID: number;
    date: string;
    meal_id: number;
    meal: Meal;
    overrides: any[];
}

export interface DailyTotals {
    date: string;
    calories: number;
    protein: number;
    fiber: number;
}
