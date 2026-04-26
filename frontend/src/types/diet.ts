export interface BaseModel {
    ID: number;
    created_at: string;
    updated_at: string;
}

export interface DietDay extends BaseModel {
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
    carbs: number;
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
    group_id?: string;
    group_label?: string;
    composite_food_id?: number | null;
}

export interface Food extends BaseModel {
    name: string;
    serving_type: string;
    serving_amount: number;
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
    variant_group_id?: number | null;
    variants?: Food[];
}

export interface SavedMeal extends BaseModel {
    name: string;
    items: SavedMealItem[];
}

export interface SavedMealItem extends BaseModel {
    saved_meal_id: number;
    food_id: number;
    food?: Food;
    amount: number;
    group_id?: string;
    group_label?: string;
    composite_food_id?: number | null;
}

export interface CompositeFoodItem extends BaseModel {
    composite_food_id: number;
    food_id: number;
    food?: Food;
    amount: number;
}

/** Recipe template; API adds entry_kind, aggregate macros for the food picker. */
export interface CompositeFood extends BaseModel {
    name: string;
    items: CompositeFoodItem[];
    entry_kind?: "composite";
    calories?: number;
    protein?: number;
    fiber?: number;
    carbs?: number;
}
