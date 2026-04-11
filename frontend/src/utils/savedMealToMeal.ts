import type { Food, Meal, MealItem, SavedMeal } from "~/types/diet";

function itemFood(it: SavedMeal["items"][number]): Food | undefined {
    return it.food ?? (it as SavedMeal["items"][number] & { Food?: Food }).Food;
}

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
                food: itemFood(it),
                amount: it.amount,
            }),
        ),
    };
}
