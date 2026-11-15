enum ValidKeyInputs {
    TAB = "TAB", //autocomplete ghost text
    ENTER = "ENTER", //autocomplete ghost text
    ESCAPE = "ESCAPE", //stop current input
}

// self note - mobile seems to only be enter
export default function onInputKeydown(event: KeyboardEvent) {

    const hasInput = trimmedItem.value.length > 0;
    const count = prefixedSuggestions.value.length;
    const hasGhost = hasInput && showGhost.value && ghostSuggestion.value;

    if (event.key === ValidKeyInputs.TAB && hasGhost) {
        event.preventDefault();
        acceptGhost();
        return;
    }
    if (event.key === "ArrowRight" && hasGhost && event.target === newItemInput.value) {
        const input = newItemInput.value;
        if (input && input.selectionStart === input.value.length && input.selectionEnd === input.value.length) {
            event.preventDefault();
            acceptGhost();
            return;
        }
    }
    if (hasInput && count > 0 && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
        event.preventDefault();
        if (event.key === "ArrowDown") {
            activeSuggestionIndex.value =
                activeSuggestionIndex.value < count - 1
                    ? activeSuggestionIndex.value + 1
                    : 0;
        } else {
            activeSuggestionIndex.value =
                activeSuggestionIndex.value > 0
                    ? activeSuggestionIndex.value - 1
                    : count - 1;
        }
        return;
    }
    if (event.key === ValidKeyInputs.ENTER) {
        event.preventDefault();
        void addItem(ghostSuggestion.value ?? undefined);
        return;
    }
    if (event.key === ValidKeyInputs.ESCAPE) {
        event.preventDefault();
        activeSuggestionIndex.value = 0;
        inputFocused.value = false;
        newItemInput.value?.blur();
    }
    return event.key;
}


