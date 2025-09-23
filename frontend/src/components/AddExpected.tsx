import { useFetchQuery } from "../hooks/useApi";
import type { DayMeal, MealPlanDay } from "../types/models";
import { formatDateShort } from "../util/dateUtil";

export default function AddExpected({ onMealSelect }: any) {
    const { data, isLoading, error } = useFetchQuery<{
        date: string;
        days: MealPlanDay[];
    }>("meal-plan-today", "/mealplan/today");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading meal plan: {error.message}</p>;

    const expectedMeals = data?.days
        ? data.days[0]?.meals?.filter((m: any) => m.status === "expected")
        : [];

    return (
        <div className="mb-12 flex flex-col gap-y-4 border-l-2 border-gray-300 pl-4">
            {data?.date && (
                <h1>Expected Meals for {formatDateShort(data?.date)}</h1>
            )}
            <div className="flex flex-row gap-x-4">
                {expectedMeals?.map((dayMeal: DayMeal) => (
                    <button
                        onClick={() => onMealSelect(dayMeal.meal)}
                        key={dayMeal.meal.id}
                        className="hover:bg-emerald-600"
                    >
                        <span className="flex flex-row gap-x-2">
                            {dayMeal.meal.name}
                        </span>
                    </button>
                ))}
            </div>
        </div>
    );
}
