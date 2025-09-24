import { createContext, useContext, type ReactNode } from "react";
import { usePopup } from "../hooks/usePopup";

type PopupContextType = {
    open: (node: ReactNode) => void;
    close: () => void;
};

const PopupContext = createContext<PopupContextType | null>(null);

export function PopupProvider({ children }: { children: ReactNode }) {
    const { open, close, Popup } = usePopup();

    return (
        <PopupContext.Provider value={{ open, close }}>
            {children}
            <Popup /> {/* renders globally once */}
        </PopupContext.Provider>
    );
}

export function useGlobalPopup() {
    const ctx = useContext(PopupContext);
    if (!ctx)
        throw new Error("useGlobalPopup must be used inside PopupProvider");
    return ctx;
}
