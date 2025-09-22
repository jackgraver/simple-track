import { useState } from "react";
import { useFetchQuery } from "../hooks/useApi";
import type { Food, Meal, MealItem } from "../types/models";
import AddExpected from "./AddExpected";

/*
Add expected meal (click once to load, click again to submit [same button])
Search other meals (modify items)
*/

function LogMeal() {
    const [currentMeal, setCurrentMeal] = useState<Meal>({
        id: 0,
        name: "",
        items: [],
        created_at: "",
        updated_at: "",
    });

    const {
        data: foods,
        isLoading: foodsLoading,
        error: foodsError,
    } = useFetchQuery<Food[]>("foods", "/foods");

    const fillExpectedMeals = (meal: any) => {
        setCurrentMeal({
            id: meal.id,
            name: meal.name,
            items: meal.items || [],
            created_at: meal.created_at,
            updated_at: meal.updated_at,
        });
    };

    const handleMealNameChange = (name: string) => {
        setCurrentMeal((prev) => ({ ...prev, name }));
    };

    const handleAddMealItem = () => {
        setCurrentMeal((prev) => ({
            ...prev,
            items: [
                ...prev.items,
                {
                    id: 0,
                    meal_id: prev.id,
                    food_id: 0,
                    amount: 0,
                    created_at: "",
                    updated_at: "",
                },
            ],
        }));
    };

    const handleUpdateMealItem = (
        index: number,
        field: keyof MealItem,
        value: any,
    ) => {
        setCurrentMeal((prev) => ({
            ...prev,
            items: prev.items.map((item, i) =>
                i === index ? { ...item, [field]: value } : item,
            ),
        }));
    };

    const handleRemoveMealItem = (index: number) => {
        setCurrentMeal((prev) => ({
            ...prev,
            items: prev.items.filter((_, i) => i !== index),
        }));
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        console.log("Submitting meal:", currentMeal);
        // TODO: Implement meal submission logic
    };

    if (foodsLoading) return <p>Loading foods...</p>;
    if (foodsError) return <p>Error loading foods: {foodsError.message}</p>;

    return (
        <div className="">
            <AddExpected onMealSelect={fillExpectedMeals} />
            <div className="border-l-2 border-gray-300 pl-4">
                <h1>Manual Log</h1>
                <form onSubmit={handleSubmit} className="mt-4">
                    <div className="mb-4 flex flex-row gap-x-4">
                        <SearchMeals
                            onMealSelect={fillExpectedMeals}
                            createNewMeal={(mealName) =>
                                console.log("Create new:", mealName)
                            }
                        />
                        <button className="mb-4 cursor-pointer rounded-lg border-none bg-red-500 text-white hover:bg-red-600">
                            X
                        </button>
                    </div>
                    <div className="mb-4">
                        {currentMeal.items.map((item, index) => (
                            <div
                                key={index}
                                className="mb-2 flex items-center gap-3"
                            >
                                <select
                                    value={item.food_id}
                                    onChange={(e) =>
                                        handleUpdateMealItem(
                                            index,
                                            "food_id",
                                            parseInt(e.target.value),
                                        )
                                    }
                                    className="flex-1 rounded border border-gray-300 p-2"
                                >
                                    <option value={0}>Select food</option>
                                    {foods?.map((food) => (
                                        <option key={food.id} value={food.id}>
                                            {food.name} ({food.calories} cal/
                                            {food.unit})
                                        </option>
                                    ))}
                                </select>
                                <input
                                    type="number"
                                    placeholder="Amount"
                                    value={item.amount}
                                    onChange={(e) =>
                                        handleUpdateMealItem(
                                            index,
                                            "amount",
                                            parseFloat(e.target.value) || 0,
                                        )
                                    }
                                    step="0.1"
                                    className="w-24 rounded border border-gray-300 p-2"
                                />
                                <button
                                    type="button"
                                    onClick={() => handleRemoveMealItem(index)}
                                    className="cursor-pointer rounded border-none bg-red-500 px-3 py-2 text-white hover:bg-red-600"
                                >
                                    ×
                                </button>
                            </div>
                        ))}
                        <button
                            type="button"
                            onClick={handleAddMealItem}
                            className=""
                        >
                            + Add Food Item
                        </button>
                    </div>
                    <div className="my-4 flex flex-row gap-x-4">
                        <span>
                            Calories:{" "}
                            {currentMeal.items.reduce(
                                (total, item) =>
                                    total +
                                    item.amount * (item.food?.calories ?? 0),
                                0,
                            )}
                        </span>
                        <span>
                            Protein:{" "}
                            {currentMeal.items.reduce(
                                (total, item) =>
                                    total +
                                    item.amount * (item.food?.protein ?? 0),
                                0,
                            )}
                        </span>
                        <span>
                            Fiber:{" "}
                            {currentMeal.items.reduce(
                                (total, item) =>
                                    total +
                                    item.amount * (item.food?.fiber ?? 0),
                                0,
                            )}
                        </span>
                    </div>
                    <button
                        type="submit"
                        disabled={currentMeal.items.length === 0}
                        className={`w-full rounded border-none p-3 text-base text-white ${
                            currentMeal.items.length === 0
                                ? "cursor-not-allowed bg-gray-400"
                                : "cursor-pointer bg-green-500 hover:bg-green-600"
                        }`}
                    >
                        Log This Meal
                    </button>
                </form>
            </div>
        </div>
    );
}

export default LogMeal;

function SearchMeals({
    onMealSelect,
    createNewMeal,
}: {
    onMealSelect: (meal: Meal) => void;
    createNewMeal?: (mealName: string) => void;
}) {
    const [searchQuery, setSearchQuery] = useState("");
    const [focused, setFocused] = useState(false);

    const {
        data: mealNames,
        isLoading,
        error,
    } = useFetchQuery<Meal[]>("meal-names", "/meals");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading meals: {error.message}</p>;

    const filteredMeals = mealNames
        ? mealNames.filter((meal) =>
              meal.name.toLowerCase().includes(searchQuery.toLowerCase()),
          )
        : [];

    return (
        <div className="mb-4">
            <input
                type="text"
                placeholder="Search meals..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full rounded border border-gray-300 p-2 text-sm"
                onFocus={() => setFocused(true)}
                onBlur={() => setFocused(false)}
            />
            {focused && filteredMeals.length > 0 && (
                <div className="absolute mt-2 max-h-32 overflow-y-auto rounded border border-gray-600 bg-gray-800">
                    {filteredMeals.map((meal, index) => (
                        <div
                            key={index}
                            onClick={() => {
                                onMealSelect(meal);
                                setSearchQuery("");
                            }}
                            className="cursor-pointer border-b border-gray-600 p-2 text-sm last:border-b-0 hover:bg-gray-600"
                        >
                            {meal.name}
                        </div>
                    ))}
                </div>
            )}
            {searchQuery && filteredMeals.length === 0 && (
                <div
                    className="absolute mt-2 rounded bg-green-600 p-2 text-sm text-white"
                    onClick={() => createNewMeal?.(searchQuery)}
                >
                    + Click to create new saved meal "{searchQuery}"
                </div>
            )}
        </div>
    );
}

//     <h3 className="text-xl font-semibold mb-4">Log a Meal</h3>

//     {/* Search Input */}
//     <div className="mb-4">
//         <input
//             type="text"
//             placeholder="Search for a meal or enter new meal name..."
//             value={searchQuery}
//             onChange={(e) => onSearchChange(e.target.value)}
//             className="w-full p-3 text-base border border-gray-300 rounded"
//         />
//     </div>

//     {/* Search Results */}
//     {searchQuery && filteredMeals.length > 0 && (
//         <div className="border border-gray-200 rounded mb-4 max-h-48 overflow-y-auto">
//             {filteredMeals.map((meal) => (
//                 <div
//                     key={meal.id}
//                     onClick={() => onMealSelect(meal)}
//                     className="p-3 cursor-pointer border-b border-gray-100 hover:bg-gray-50"
//                 >
//                     <div className="font-semibold">{meal.name}</div>
//                     <div className="text-sm text-gray-600">
//                         {meal.items.length} items
//                     </div>
//                 </div>
//             ))}
//         </div>
//     )}

//     {/* Meal Form */}
//     <form onSubmit={onSubmit}>
//         <div className="mb-4">
//             <input
//                 type="text"
//                 placeholder="Meal name"
//                 value={currentMeal.name}
//                 onChange={(e) => onMealNameChange(e.target.value)}
//                 required
//                 className="w-full p-2 text-base border border-gray-300 rounded"
//             />
//         </div>

//         {/* Meal Items */}
//         <div className="mb-4">
//             <button
//                 type="button"
//                 onClick={onAddMealItem}
//                 className="px-4 py-2 bg-blue-500 text-white border-none rounded cursor-pointer mb-3 hover:bg-blue-600"
//             >
//                 + Add Food Item
//             </button>

//             {currentMeal.items.map((item, index) => (
//                 <div
//                     key={index}
//                     className="flex gap-3 mb-2 items-center"
//                 >
//                     <select
//                         value={item.food_id}
//                         onChange={(e) =>
//                             onUpdateMealItem(
//                                 index,
//                                 "food_id",
//                                 parseInt(e.target.value)
//                             )
//                         }
//                         className="flex-1 p-2 border border-gray-300 rounded"
//                     >
//                         <option value={0}>Select food</option>
//                         {foods.map((food) => (
//                             <option key={food.id} value={food.id}>
//                                 {food.name} ({food.calories} cal/
//                                 {food.unit})
//                             </option>
//                         ))}
//                     </select>
//                     <input
//                         type="number"
//                         placeholder="Amount"
//                         value={item.amount}
//                         onChange={(e) =>
//                             onUpdateMealItem(
//                                 index,
//                                 "amount",
//                                 parseFloat(e.target.value) || 0
//                             )
//                         }
//                         step="0.1"
//                         className="w-24 p-2 border border-gray-300 rounded"
//                     />
//                     <button
//                         type="button"
//                         onClick={() => onRemoveMealItem(index)}
//                         className="px-3 py-2 bg-red-500 text-white border-none rounded cursor-pointer hover:bg-red-600"
//                     >
//                         ×
//                     </button>
//                 </div>
//             ))}
//         </div>

//         {/* Save as new meal checkbox */}
//         <div className="mb-4">
//             <label className="flex items-center gap-2">
//                 <input
//                     type="checkbox"
//                     checked={saveAsNewMeal}
//                     onChange={(e) =>
//                         onSaveAsNewMealChange(e.target.checked)
//                     }
//                 />
//                 Save as new meal template
//             </label>
//         </div>

//         {/* Submit Button */}
//         <button
//             type="submit"
//             disabled={currentMeal.items.length === 0}
//             className={`w-full p-3 text-base text-white border-none rounded ${
//                 currentMeal.items.length === 0
//                     ? "bg-gray-400 cursor-not-allowed"
//                     : "bg-green-500 cursor-pointer hover:bg-green-600"
//             }`}
//         >
//             Log This Meal
//         </button>
//     </form>
