import { useFetchQuery } from "../hooks/useApi";
import type { Food } from "../types/models";

function AllFoods() {
    const { data, isLoading, error } = useFetchQuery<Food[]>("foods", "/foods");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading foods: {error.message}</p>;

    // const filteredMeals = meals.filter((meal) =>
    //     meal.name.toLowerCase().includes(searchQuery.toLowerCase())
    // );
    const foods = data || [];

    return (
        <div className="border border-gray-300 p-4 mb-5 rounded-lg">
            <h3 className="text-xl font-semibold mb-4">
                All Foods ({foods.length})
            </h3>
            {foods.length === 0 ? (
                <p className="text-gray-600 italic">No foods in database yet</p>
            ) : (
                <div className="flex flex-col gap-3">
                    {foods.map((food) => (
                        <div
                            key={food.id}
                            className="p-3 rounded border border-gray-200 bg-gray-50"
                        >
                            <div className="font-semibold mb-2">
                                {food.name}
                            </div>
                            <div className="text-sm text-gray-600 space-y-1">
                                <div>Unit: {food.unit}</div>
                                <div>
                                    Calories: {food.calories} per {food.unit}
                                </div>
                                <div>
                                    Protein: {food.protein}g per {food.unit}
                                </div>
                                <div>
                                    Carbs: {food.carbs}g per {food.unit}
                                </div>
                                <div>
                                    Fat: {food.fat}g per {food.unit}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

export default AllFoods;
