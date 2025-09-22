import { useState, useEffect } from "react";
import "./App.css";
import type { Meal, MealLog, DailyTotals } from "./types/models";
import TodayProgress from "./components/TodayProgress";
import LogMeal from "./components/LogMeal";
import RecentMeals from "./components/RecentMeals";
import AllFoods from "./components/AllFoods";
import AllMeals from "./components/AllMeals";
import MainLayout from "./layouts/MainLayout";
import ManageMealPlan from "./components/ManageMealPlan";

const API_BASE = "http://localhost:8080/api";

function App() {
    // // State
    // const [mealLogs, setMealLogs] = useState<MealLog[]>([]);

    // // Search and form states
    // const [searchQuery, setSearchQuery] = useState("");
    // const [selectedMeal, setSelectedMeal] = useState<Meal | null>(null);
    // const [currentMeal, setCurrentMeal] = useState({
    //     name: "",
    //     items: [] as { food_id: number; amount: number }[],
    // });
    // const [saveAsNewMeal, setSaveAsNewMeal] = useState(false);

    // // Load data on component mount
    // useEffect(() => {
    //     loadMealLogs();
    // }, []);

    // const loadMealLogs = async () => {
    //     try {
    //         const response = await fetch(`${API_BASE}/meal-logs`);
    //         const data = await response.json();
    //         setMealLogs(data);
    //     } catch (error) {
    //         console.error("Error loading meal logs:", error);
    //     }
    // };

    // const addMeal = async (mealData: any) => {
    //     try {
    //         const response = await fetch(`${API_BASE}/meals`, {
    //             method: "POST",
    //             headers: { "Content-Type": "application/json" },
    //             body: JSON.stringify(mealData),
    //         });
    //         if (response.ok) {
    //             // loadMeals();
    //             return true;
    //         }
    //     } catch (error) {
    //         console.error("Error adding meal:", error);
    //     }
    //     return false;
    // };

    // const logMeal = async (mealData: any) => {
    //     try {
    //         const response = await fetch(`${API_BASE}/meal-logs`, {
    //             method: "POST",
    //             headers: { "Content-Type": "application/json" },
    //             body: JSON.stringify(mealData),
    //         });
    //         if (response.ok) {
    //             loadMealLogs();
    //             // loadDailyTotals();
    //             return true;
    //         }
    //     } catch (error) {
    //         console.error("Error logging meal:", error);
    //     }
    //     return false;
    // };

    // // Handle meal selection
    // const handleMealSelect = (meal: Meal) => {
    //     setSelectedMeal(meal);
    //     setCurrentMeal({
    //         name: meal.name,
    //         items: meal.items.map((item) => ({
    //             food_id: item.food_id,
    //             amount: item.amount,
    //         })),
    //     });
    //     setSearchQuery(meal.name);
    // };

    // // Handle search input change
    // const handleSearchChange = (query: string) => {
    //     setSearchQuery(query);
    //     if (query === "") {
    //         setSelectedMeal(null);
    //         setCurrentMeal({ name: "", items: [] });
    //     }
    // };

    // // Add food item to current meal
    // const addMealItem = () => {
    //     setCurrentMeal((prev) => ({
    //         ...prev,
    //         items: [...prev.items, { food_id: 0, amount: 0 }],
    //     }));
    // };

    // // Update meal item
    // const updateMealItem = (index: number, field: string, value: any) => {
    //     setCurrentMeal((prev) => ({
    //         ...prev,
    //         items: prev.items.map((item, i) =>
    //             i === index ? { ...item, [field]: value } : item
    //         ),
    //     }));
    // };

    // // Remove meal item
    // const removeMealItem = (index: number) => {
    //     setCurrentMeal((prev) => ({
    //         ...prev,
    //         items: prev.items.filter((_, i) => i !== index),
    //     }));
    // };

    // // Submit meal
    // const handleSubmitMeal = async (e: React.FormEvent) => {
    //     e.preventDefault();

    //     if (currentMeal.items.length === 0) return;

    //     // If saving as new meal, create it first
    //     let mealId = selectedMeal?.id;
    //     if (saveAsNewMeal || !selectedMeal) {
    //         const success = await addMeal(currentMeal);
    //         if (!success) return;
    //         // Reload meals to get the new meal ID
    //         // await loadMeals();
    //         // Find the newly created meal
    //         // const newMeal = meals.find((m) => m.name === currentMeal.name);
    //         // mealId = newMeal?.id;
    //     }

    //     if (mealId) {
    //         const success = await logMeal({ meal_id: mealId });
    //         if (success) {
    //             // Reset form
    //             setCurrentMeal({ name: "", items: [] });
    //             setSelectedMeal(null);
    //             setSearchQuery("");
    //             setSaveAsNewMeal(false);
    //         }
    //     }
    // };

    return (
        // <MainLayout>
        <div className="flex flex-col gap-x-12">
            <div className="flex w-1/2 justify-center">
                <div className="flex flex-col">
                    <LogMeal />
                    <TodayProgress />
                </div>
            </div>
            <ManageMealPlan />
        </div>

        // </MainLayout>
    );
}

export default App;
{
    /* <LogMeal
                searchQuery={searchQuery}
                filteredMeals={[]}
                currentMeal={currentMeal}
                saveAsNewMeal={saveAsNewMeal}
                foods={[]}
                onSearchChange={handleSearchChange}
                onMealSelect={handleMealSelect}
                onMealNameChange={(name) =>
                    setCurrentMeal((prev) => ({ ...prev, name }))
                }
                onAddMealItem={addMealItem}
                onUpdateMealItem={updateMealItem}
                onRemoveMealItem={removeMealItem}
                onSaveAsNewMealChange={setSaveAsNewMeal}
                onSubmit={handleSubmitMeal}
            /> */
}
//   <div className="px-4 sm:px-0">
//     <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
//         <div className="space-y-6">

//             <LogMeal
//                 searchQuery={searchQuery}
//                 filteredMeals={[]}
//                 currentMeal={currentMeal}
//                 saveAsNewMeal={saveAsNewMeal}
//                 foods={[]}
//                 onSearchChange={handleSearchChange}
//                 onMealSelect={handleMealSelect}
//                 onMealNameChange={(name) =>
//                     setCurrentMeal((prev) => ({ ...prev, name }))
//                 }
//                 onAddMealItem={addMealItem}
//                 onUpdateMealItem={updateMealItem}
//                 onRemoveMealItem={removeMealItem}
//                 onSaveAsNewMealChange={setSaveAsNewMeal}
//                 onSubmit={handleSubmitMeal}
//             />
//             <RecentMeals mealLogs={mealLogs} />
//         </div>
//        <AllFoods />
//         <AllMeals />
//     </div>
// </div>
