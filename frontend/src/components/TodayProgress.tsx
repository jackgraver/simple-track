import { useFetchQuery } from "../hooks/useApi";
import type { DailyTotals } from "../types/models";
import { CircularProgressbar } from "react-circular-progressbar";
import "react-circular-progressbar/dist/styles.css";

type DailyGoals = {
    date: string;
    calories: number;
    protein: number;
    fiber: number;
};

function TodayProgress() {
    const dailyTotalsQuery = {
        date: "",
        calories: 2000,
        protein: 100,
        fiber: 30,
    };
    // useFetchQuery<DailyTotals>(
    //     "daily-totals",
    //     "/daily-totals",
    // );
    const dailyGoalsQuery = useFetchQuery<DailyGoals>(
        "today-goals",
        "mealplan/goals/today",
    );

    // if (dailyTotalsQuery.isLoading || dailyGoalsQuery.isLoading) {
    if (dailyGoalsQuery.isLoading) {
        return <p>Loading...</p>;
    }

    if (dailyGoalsQuery.error) {
        return <p>Error loading data</p>;
    }

    const totals = dailyTotalsQuery;
    const goals = dailyGoalsQuery.data;
    console.log(goals);
    const percentage =
        totals && goals
            ? {
                  calories: Math.round(
                      (totals.calories / goals.calories) * 100,
                  ),
                  protein: Math.round((totals.protein / goals.protein) * 100),
                  fiber: Math.round((0 / goals.fiber) * 100),
              }
            : {
                  calories: 0,
                  protein: 0,
                  fiber: 0,
              };

    return (
        <div className="mb-8 mt-4 border-l-2 border-gray-300 pl-4">
            <h3 className="mb-4 text-xl font-semibold">Today's Progress</h3>
            <div className="flex flex-row gap-x-4">
                <div style={{ width: 128, height: 128 }}>
                    <CircularProgressbar
                        value={percentage.calories}
                        text={`${percentage.calories}%`}
                    />
                </div>
                <div style={{ width: 128, height: 128 }}>
                    <CircularProgressbar
                        value={percentage.protein}
                        text={`${percentage.protein}%`}
                    />
                </div>
                <div style={{ width: 128, height: 128 }}>
                    <CircularProgressbar
                        value={percentage.fiber}
                        text={`${percentage.fiber}%`}
                    />
                </div>
            </div>
        </div>
    );
}

export default TodayProgress;
