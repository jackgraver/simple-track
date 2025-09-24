import { useState } from "react";
import { api, useFetchQuery } from "../hooks/useApi";
import type { DayMeal, Food, Meal, MealItem } from "../types/models";
import AddExpected from "./AddExpected";
import { useMutation } from "@tanstack/react-query";

/*
Add expected meal (click once to load, click again to submit [same button])
Search other meals (modify items)
*/

function LogMeal() {
    const [currentMeal, setCurrentMeal] = useState<Meal>({
        ID: -1,
        name: "",
        items: [],
        created_at: "",
        updated_at: "",
    });

    const logMealMutation = useMutation({
        mutationFn: async (mealData: {
            meal_id: number;
            status: "expected" | "actual";
        }) => {
            const res = await api.post("/mealplan/meal/log", mealData);
            return res.data;
        },
        onSuccess: (data) => {
            console.log("Created:", data);
        },
    });

    const {
        data: foods,
        isLoading: foodsLoading,
        error: foodsError,
    } = useFetchQuery<Food[]>("food-all", "mealplan/food/all");

    const fillExpectedMeals = (meal: Meal) => {
        setCurrentMeal({
            ID: meal.ID,
            name: meal.name,
            items: meal.items || [],
            created_at: meal.created_at,
            updated_at: meal.updated_at,
        });
    };

    const clearCurrentMeal = () => {
        setCurrentMeal({
            ID: -1,
            name: "",
            items: [],
            created_at: "",
            updated_at: "",
        });
    };

    const handleAddMealItem = () => {
        setCurrentMeal((prev) => ({
            ...prev,
            items: [
                ...prev.items,
                {
                    ID: 0,
                    meal_id: prev.ID,
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

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        logMealMutation.mutate({
            meal_id: currentMeal.ID,
            status: "actual",
        });

        // const res = await fetch("http://localhost:8080/mealplan/meal/log", {
        //     method: "POST",
        //     headers: { "Content-Type": "application/json" },
        //     body: JSON.stringify({
        //         meal_id: currentMeal.ID,
        //         status: "actual",
        //     }),
        // });

        // if (!res.ok) {
        //     throw new Error("Failed to add actual meal");
        // }

        // const data = await res.json();
        // return data;
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
                        <button
                            className="mb-4 cursor-pointer rounded-lg border-none bg-red-500 text-white hover:bg-red-600"
                            onCanPlay={() => clearCurrentMeal()}
                        >
                            X
                        </button>
                    </div>
                    {currentMeal.ID > 0 && (
                        <>
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
                                            <option value={0}>
                                                Select food
                                            </option>
                                            {foods?.map((food) => (
                                                <option
                                                    key={food.ID}
                                                    value={food.ID}
                                                >
                                                    {food.name} ({food.calories}{" "}
                                                    cal/
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
                                                    parseFloat(
                                                        e.target.value,
                                                    ) || 0,
                                                )
                                            }
                                            step="0.1"
                                            className="w-24 rounded border border-gray-300 p-2"
                                        />
                                        <button
                                            type="button"
                                            onClick={() =>
                                                handleRemoveMealItem(index)
                                            }
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
                                            item.amount *
                                                (item.food?.calories ?? 0),
                                        0,
                                    )}
                                </span>
                                <span>
                                    Protein:{" "}
                                    {currentMeal.items.reduce(
                                        (total, item) =>
                                            total +
                                            item.amount *
                                                (item.food?.protein ?? 0),
                                        0,
                                    )}
                                </span>
                                <span>
                                    Fiber:{" "}
                                    {currentMeal.items.reduce(
                                        (total, item) =>
                                            total +
                                            item.amount *
                                                (item.food?.fiber ?? 0),
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
                        </>
                    )}
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
    } = useFetchQuery<Meal[]>("meal-names", "mealplan/meal/all");

    if (isLoading) return <p>Loading...</p>;
    if (error) return <p>Error loading meals: {error.message}</p>;

    const filteredMeals = mealNames
        ? mealNames.filter((meal) =>
              meal.name.toLowerCase().includes(searchQuery.toLowerCase()),
          )
        : [];

    console.log(mealNames);

    return (
        <div className="mb-4">
            <input
                type="text"
                placeholder="Search meals..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full rounded border border-gray-300 p-2 text-sm"
                onFocus={() => setFocused(true)}
                // onBlur={() => {
                //     setTimeout(() => setFocused(false), 150);
                // }}
            />
            {focused && filteredMeals.length > 0 && (
                <div className="absolute mt-2 max-h-32 overflow-y-auto rounded border border-gray-600 bg-gray-800">
                    {filteredMeals.map((meal, index) => (
                        <div
                            key={index}
                            onClick={() => {
                                setFocused(false);
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
