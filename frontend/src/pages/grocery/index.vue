<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from "vue";
import {
    useCompleteGroceryItem,
    useCreateGroceryItem,
    useGroceryItems,
    useGrocerySuggestions,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";

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
    <div class="flex w-full min-w-0 flex-col gap-4">
        <h1 class="m-0 text-lg font-semibold text-textPrimary">Grocery List</h1>
        <div v-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load grocery list" }}
        </div>
        <div v-else-if="isPending" class="text-sm text-textSecondary">
            Loading…
        </div>
        <section v-else class="flex flex-col">
            <div class="flex flex-col">
                <div
                    v-for="item in displayItems"
                    :key="item.ID"
                    class="grocery-row-collapse grid transition-[grid-template-rows] duration-500 ease-in-out"
                    :class="
                        isRemoving(item.ID)
                            ? 'grid-rows-[0fr]'
                            : 'grid-rows-[1fr]'
                    "
                >
                    <div class="min-h-0 overflow-hidden">
                        <div
                            class="grocery-row flex items-center gap-3 py-1.5 text-sm transition-opacity duration-500 ease-out"
                            :class="{
                                'grocery-row--completing': isCompleting(item.ID),
                                'opacity-0': isRemoving(item.ID),
                            }"
                        >
                            <input
                                type="checkbox"
                                class="m-0 size-4 shrink-0 cursor-pointer"
                                :aria-label="`Complete ${item.name}`"
                                :checked="isCompleting(item.ID)"
                                @change="onItemCheck(item.ID, $event)"
                            />
                            <span
                                class="grocery-row__label min-w-0 flex-1 truncate"
                                >{{ item.name }}</span
                            >
                        </div>
                    </div>
                </div>
            </div>
            <div class="flex items-center gap-3 py-1.5">
                <input
                    type="checkbox"
                    class="m-0 size-4 shrink-0 pointer-events-none opacity-30"
                    disabled
                    aria-hidden="true"
                />
                <div class="grocery-input-wrap relative min-w-0 flex-1">
                    <div
                        aria-hidden="true"
                        class="grocery-input-mirror pointer-events-none absolute inset-0 overflow-hidden whitespace-pre"
                    >
                        <span class="text-textPrimary">{{ newItem }}</span
                        ><span v-if="showGhost" class="text-textSecondary/45">{{
                            ghostSuffix
                        }}</span>
                    </div>
                    <input
                        ref="newItemInput"
                        v-model="newItem"
                        type="text"
                        autocomplete="off"
                        aria-label="Add grocery item"
                        placeholder="Add item…"
                        class="grocery-input relative w-full min-w-0 outline-none"
                        @focus="inputFocused = true"
                        @blur="onInputBlur"
                        @input="onInputInput"
                        @keydown="onInputKeydown"
                    />
                </div>
            </div>
        </section>
    </div>
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
