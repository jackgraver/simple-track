import type { Meal, MealItem, SavedMeal } from "~/types/diet";

export function savedMealToMeal(s: SavedMeal): Meal {
    return {
        ID: 0,
        created_at: "",
        updated_at: "",
        name: s.name,
        items: s.items.map(
            (it): MealItem => ({
                ID: 0,
                created_at: "",
                updated_at: "",
                meal_id: 0,
                food_id: it.food_id,
                food: it.food,
                amount: it.amount,
            }),
        ),
    };
}
