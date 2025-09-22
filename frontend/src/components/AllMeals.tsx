import { useFetchQuery } from "../hooks/useApi";
import type { Meal } from "../types/models";

function AllMeals() {
    const { data, isLoading, error } = useFetchQuery<Meal[]>("meals", "/meals");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading meals: {error.message}</p>;

    const meals = data || [];

    return (
        <div className="border border-gray-300 p-4 rounded-lg">
            <h3 className="text-xl font-semibold mb-4">
                All Meals ({meals.length})
            </h3>
            {meals.length === 0 ? (
                <p className="text-gray-600 italic">No meals in database yet</p>
            ) : (
                <div className="flex flex-col gap-4">
                    {meals.map((meal) => (
                        <div
                            key={meal.id}
                            className="p-4 rounded border border-gray-200 bg-gray-50"
                        >
                            <div className="font-semibold mb-2 text-lg">
                                {meal.name}
                            </div>
                            <div className="text-sm text-gray-600 mb-2">
                                Created:{" "}
                                {new Date(meal.created_at).toLocaleDateString()}
                            </div>
                            <div className="mb-2">
                                <strong>Items ({meal.items.length}):</strong>
                            </div>
                            <div className="text-sm space-y-1">
                                {meal.items.map((item, index) => (
                                    <div
                                        key={index}
                                        className="p-2 rounded border border-gray-300 bg-white"
                                    >
                                        <span className="font-medium">
                                            {item.amount}
                                            {item.food?.unit}
                                        </span>{" "}
                                        {item.food?.name}
                                        {item.food && (
                                            <span className="text-gray-600 text-xs">
                                                {" "}
                                                (
                                                {item.food.calories *
                                                    item.amount}{" "}
                                                cal)
                                            </span>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

export default AllMeals;
