import type { MealLog } from "../types/models";

interface RecentMealsProps {
    mealLogs: MealLog[];
}

function RecentMeals({ mealLogs }: RecentMealsProps) {
    return (
        <div className="border border-gray-300 p-4 rounded-lg">
            <h3 className="text-xl font-semibold mb-4">
                Recent Meals ({mealLogs.length})
            </h3>
            {mealLogs.length === 0 ? (
                <p className="text-gray-600 italic">
                    No meals logged yet today
                </p>
            ) : (
                <div className="space-y-2">
                    {mealLogs.map((log) => (
                        <div
                            key={log.id}
                            className="p-2 rounded border border-gray-200 bg-gray-50"
                        >
                            <div className="font-semibold">{log.meal.name}</div>
                            <div className="text-sm text-gray-600">
                                {new Date(log.date).toLocaleDateString()} at{" "}
                                {new Date(log.date).toLocaleTimeString()}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

export default RecentMeals;
