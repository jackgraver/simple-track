import "./App.css";
import TodayProgress from "./components/TodayProgress";
import LogMeal from "./components/LogMeal";
import ManageMealPlan from "./components/ManageMealPlan";
import { usePopup } from "./hooks/usePopup";

function App() {
    const popup = usePopup();

    return (
        <div className="flex flex-col gap-x-12">
            <div className="flex w-1/2 justify-center">
                <div className="flex flex-col">
                    <LogMeal />
                    <TodayProgress />
                </div>
            </div>
            <ManageMealPlan />
        </div>
    );
}

export default App;
