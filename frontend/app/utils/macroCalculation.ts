import type { MealPlanDay } from "~/types/models";

export function totalCaloriesEaten(day: MealPlanDay): number {
    if (!day.meals) return 0;
    return parseFloat(
        day.meals
            .filter((dm) => dm.status === "actual") // only actual meals
            .reduce((sum, dm) => {
                // sum calories for each MealItem in the meal
                const mealCalories = dm.meal.items.reduce((mealSum, item) => {
                    if (!item.food) return mealSum;
                    return mealSum + item.food.calories * item.amount;
                }, 0);
                return sum + mealCalories;
            }, 0)
            .toFixed(2),
    );
}
export function totalProteinEaten(day: MealPlanDay): number {
    if (!day.meals) return 0;
    return parseFloat(
        day.meals
            .filter((dm) => dm.status === "actual") // only actual meals
            .reduce((sum, dm) => {
                // sum calories for each MealItem in the meal
                const mealProtein = dm.meal.items.reduce((mealSum, item) => {
                    if (!item.food) return mealSum;
                    return mealSum + item.food.protein * item.amount;
                }, 0);
                return sum + mealProtein;
            }, 0)
            .toFixed(2),
    );
}
export function totalFiberEaten(day: MealPlanDay): number {
    if (!day.meals) return 0;
    return parseFloat(
        day.meals
            .filter((dm) => dm.status === "actual") // only actual meals
            .reduce((sum, dm) => {
                // sum calories for each MealItem in the meal
                const mealFiber = dm.meal.items.reduce((mealSum, item) => {
                    if (!item.food) return mealSum;
                    return mealSum + item.food.fiber * item.amount;
                }, 0);
                return sum + mealFiber;
            }, 0)
            .toFixed(2),
    );
}
