import { X } from "lucide-react";
import { useState, type ReactNode } from "react";

export function usePopup() {
    const [isOpen, setIsOpen] = useState(true);
    const [content, setContent] = useState<ReactNode>(null);

    function open(node: ReactNode) {
        setContent(node);
        setIsOpen(true);
    }
    function close() {
        setIsOpen(false);
        setContent(null);
    }

    const Popup = () =>
        isOpen ? (
            <div className="absolute left-1/2 top-1/4 bg-gray-500">
                <div className="flex flex-row border-b border-gray-200">
                    <h1 className="flex-1 p-2 text-2xl font-semibold">Title</h1>
                    <button onClick={() => close()} className="bg-transparent">
                        <X className="h-6 w-6" />
                    </button>
                </div>
                <div>{content}</div>
            </div>
        ) : null;

    return { open, close, Popup };
}
