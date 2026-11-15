<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from "vue";
import {
    useCompleteGroceryItem,
    useCreateGroceryItem,
    useGroceryItems,
    useGrocerySuggestions,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";

import GroceryList from "./GroceryList.vue"

const COMPLETE_DELAY_MS = 1200;
const COLLAPSE_MS = 500;

const newItem = ref("");
const completingIds = ref<Set<number>>(new Set());
const removingIds = ref<Set<number>>(new Set());
const hiddenIds = ref<Set<number>>(new Set());
const inputFocused = ref(false);
const activeSuggestionIndex = ref(0);
const newItemInput = useTemplateRef<HTMLInputElement>("newItemInput");

const { data, isError, isPending, error } = useGroceryItems();
const suggestionsQuery = useGrocerySuggestions(newItem);
const createMutation = useCreateGroceryItem();
const completeMutation = useCompleteGroceryItem();

const items = computed(() => data.value ?? []);
const displayItems = computed(() =>
    items.value.filter((item) => !hiddenIds.value.has(item.ID)),
);

watch(items, (list) => {
    const present = new Set(list.map((item) => item.ID));
    const next = new Set(
        [...hiddenIds.value].filter((id) => present.has(id)),
    );
    if (next.size !== hiddenIds.value.size) {
        hiddenIds.value = next;
    }
});
const trimmedItem = computed(() => newItem.value.trim());
const suggestions = computed(() =>
    trimmedItem.value.length > 0 ? (suggestionsQuery.data.value ?? []) : [],
);

const prefixedSuggestions = computed(() => {
    const query = trimmedItem.value;
    if (!query) return [];
    const lower = query.toLowerCase();
    return suggestions.value.filter((s) => s.toLowerCase().startsWith(lower));
});

const ghostSuggestion = computed(() => {
    if (
        !inputFocused.value ||
        trimmedItem.value.length === 0 ||
        prefixedSuggestions.value.length === 0
    ) {
        return null;
    }
    const index = Math.min(
        Math.max(activeSuggestionIndex.value, 0),
        prefixedSuggestions.value.length - 1,
    );
    return prefixedSuggestions.value[index] ?? null;
});

const ghostSuffix = computed(() => {
    const suggestion = ghostSuggestion.value;
    if (!suggestion) return "";
    return suggestion.slice(newItem.value.length);
});

const showGhost = computed(
    () => inputFocused.value && ghostSuffix.value.length > 0,
);

function isCompleting(id: number) {
    return completingIds.value.has(id);
}

function isRemoving(id: number) {
    return removingIds.value.has(id);
}

function updateIdSet(
    target: "completing" | "removing",
    id: number,
    value: boolean,
) {
    const source =
        target === "completing" ? completingIds : removingIds;
    const next = new Set(source.value);
    if (value) {
        next.add(id);
    } else {
        next.delete(id);
    }
    source.value = next;
}

async function focusNewItemInput() {
    await nextTick();
    newItemInput.value?.focus();
}

async function addItem(name?: string) {
    const value = (name ?? trimmedItem.value).trim();
    if (!value || createMutation.isPending.value) return;
    try {
        await createMutation.mutateAsync({ name: value });
        newItem.value = "";
        activeSuggestionIndex.value = 0;
        await focusNewItemInput();
    } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to add item";
        toast.push(message, "error");
    }
}

function onItemCheck(id: number, event: Event) {
    const checked = (event.target as HTMLInputElement).checked;
    if (checked) {
        void completeItem(id);
    }
}

function hideItem(id: number) {
    hiddenIds.value = new Set([...hiddenIds.value, id]);
}

function unhideItem(id: number) {
    const next = new Set(hiddenIds.value);
    next.delete(id);
    hiddenIds.value = next;
}

function clearItemState(id: number) {
    updateIdSet("completing", id, false);
    updateIdSet("removing", id, false);
    unhideItem(id);
}

async function completeItem(id: number) {
    if (isCompleting(id) || isRemoving(id)) return;
    updateIdSet("completing", id, true);
    const persistPromise = completeMutation.mutateAsync(id);
    await new Promise((resolve) => setTimeout(resolve, COMPLETE_DELAY_MS));
    try {
        await persistPromise;
    } catch (err) {
        clearItemState(id);
        const message =
            err instanceof Error ? err.message : "Failed to complete item";
        toast.push(message, "error");
        return;
    }
    updateIdSet("removing", id, true);
    await new Promise((resolve) => setTimeout(resolve, COLLAPSE_MS));
    hideItem(id);
    updateIdSet("completing", id, false);
    updateIdSet("removing", id, false);
}

function acceptGhost() {
    if (!ghostSuggestion.value) return;
    newItem.value = ghostSuggestion.value;
}

function onInputBlur() {
    window.setTimeout(() => {
        inputFocused.value = false;
        activeSuggestionIndex.value = 0;
    }, 120);
}

function onInputInput() {
    activeSuggestionIndex.value = 0;
}

function onInputKeydown(event: KeyboardEvent) {
    const hasInput = trimmedItem.value.length > 0;
    const count = prefixedSuggestions.value.length;
    const hasGhost = hasInput && showGhost.value && ghostSuggestion.value;

    if (event.key === "Tab" && hasGhost) {
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
    if (event.key === "Enter") {
        event.preventDefault();
        void addItem(ghostSuggestion.value ?? undefined);
        return;
    }
    if (event.key === "Escape") {
        event.preventDefault();
        activeSuggestionIndex.value = 0;
        inputFocused.value = false;
        newItemInput.value?.blur();
    }
}
</script>

<template>
    <GroceryList />
</template>

<style scoped>
.grocery-input-wrap {
    min-height: 1.25rem;
}

.grocery-input-mirror,
.grocery-input {
    font: inherit;
    font-size: 0.875rem;
    line-height: 1.25rem;
    padding: 0;
    margin: 0;
    border: none;
}

.grocery-input {
    background: transparent !important;
    color: transparent !important;
    caret-color: rgb(218, 218, 218);
}

.grocery-input::placeholder {
    color: rgb(140, 140, 140);
}

.grocery-row {
    color: var(--color-text-primary, rgb(218, 218, 218));
    opacity: 1;
    transition:
        color 0.7s ease,
        opacity 0.7s ease;
}

.grocery-row__label {
    transition:
        color 0.7s ease,
        opacity 0.7s ease,
        text-decoration-color 0.7s ease;
    text-decoration: line-through transparent;
    text-decoration-thickness: 1px;
}

.grocery-row--completing {
    color: rgb(140, 140, 140);
    opacity: 0.55;
}

.grocery-row--completing .grocery-row__label {
    text-decoration-color: rgb(140, 140, 140);
}
</style>
