import type { Meal, MealItem } from "~/types/diet";

export function cloneMealForNewLog(source: Meal): Meal {
    return {
        ...source,
        ID: 0,
        created_at: "",
        updated_at: "",
        items: source.items.map(
            (it): MealItem => ({
                ...it,
                ID: 0,
                meal_id: 0,
                food_id: it.food_id,
                food: it.food,
                amount: it.amount,
            }),
        ),
    };
}
