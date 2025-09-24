import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { PopupProvider } from "./context/PopupContext.tsx";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
    <QueryClientProvider client={queryClient}>
        <PopupProvider>
            <StrictMode>
                <App />
            </StrictMode>
        </PopupProvider>
    </QueryClientProvider>,
);
