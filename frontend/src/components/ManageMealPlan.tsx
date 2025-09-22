import { useFetchQuery } from "../hooks/useApi";
import type { Meal, MealItem, MealPlanDay } from "../types/models";
import { formatDateShort, isSameDay } from "../util/dateUtil";

export default function ManageMealPlan() {
    const { data, isLoading, error } = useFetchQuery<{
        today: string;
        days: MealPlanDay[];
    }>("meal-plan-days", "/meal-plan-days");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading meal plan days: {error.message}</p>;

    const days = data?.days;
    const today = data?.today;

    console.log(today);

    return (
        <div className="flex flex-row gap-x-4 pt-12">
            {days?.map((day, i) => {
                const expectedMeals = day.meals?.filter(
                    (m) => m.status === "expected",
                );
                const actualMeals = day.meals?.filter(
                    (m) => m.status === "actual",
                );

                return (
                    <div
                        key={i}
                        className={`mb-4 rounded-lg border border-gray-300 p-4 ${today && isSameDay(today, day.date) ? "bg-gray-600" : ""}`}
                    >
                        <p className="font-bold">{formatDateShort(day.date)}</p>
                        <span className="flex flex-row gap-x-2">
                            <p className="text-orange-400">
                                C {day.goals.calories}
                            </p>
                            <p className="text-blue-500">
                                P {day.goals.protein}g
                            </p>
                            <p className="text-emerald-500">
                                F {day.goals.fiber}g
                            </p>
                        </span>

                        <div>
                            <h4>Expected</h4>
                            {expectedMeals?.map((dayMeal) => (
                                <p key={dayMeal.meal.id}>{dayMeal.meal.name}</p>
                            ))}
                        </div>

                        <div>
                            <h4>Actual</h4>
                            {actualMeals?.map((dayMeal) => (
                                <p key={dayMeal.meal.id}>{dayMeal.meal.name}</p>
                            ))}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
