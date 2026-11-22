<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from "vue";
import {
    useCompleteGroceryItem,
    useCreateGroceryItem,
    useGroceryItems,
    useGrocerySuggestions,
} from "~/api/tracking/queries";
import { newClientId } from "~/api/ids";
import { isTempGroceryId } from "~/api/tracking/groceryCache";
import { ClipboardList, ShoppingBasket } from "lucide-vue-next";
import {
    Button,
    Card,
    Checkbox,
    EmptyState,
    ErrorState,
    Input,
    List,
    ListRow,
    PageHeader,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { openAddListDialog } from "./dialog/useListDialog";

const COMPLETE_DELAY_MS = 1200;

const newItem = ref("");
const completingIds = ref<Set<number>>(new Set());
const hiddenIds = ref<Set<number>>(new Set());
const inputFocused = ref(false);
const activeSuggestionIndex = ref(0);
const addWrap = useTemplateRef<HTMLElement>("addWrap");
const newItemInput = computed(() => addWrap.value?.querySelector("input") ?? null);

const { data, isError, isPending, error, refetch } = useGroceryItems();
const showSkeleton = useDelayedFlag(isPending);
const suggestionsQuery = useGrocerySuggestions(newItem);
const createMutation = useCreateGroceryItem();
const completeMutation = useCompleteGroceryItem();

const items = computed(() => data.value ?? []);
// The list refetches as soon as the check is saved, which would drop the row before the strike-through
// has been seen. Remember where each row was and keep completing rows in place until they're hidden.
const lastSeen = ref(new Map<number, { item: (typeof items.value)[number]; index: number }>());
watch(
    items,
    (list) => {
        const next = new Map(lastSeen.value);
        list.forEach((item, index) => next.set(item.ID, { item, index }));
        lastSeen.value = next;
    },
    { immediate: true },
);
const displayItems = computed(() => {
    const list = items.value.filter((item) => !hiddenIds.value.has(item.ID));
    const present = new Set(list.map((item) => item.ID));
    for (const id of completingIds.value) {
        const seen = lastSeen.value.get(id);
        if (!present.has(id) && seen) list.splice(Math.min(seen.index, list.length), 0, seen.item);
    }
    return list;
});

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

const activeSuggestion = computed(() => {
    if (!inputFocused.value || prefixedSuggestions.value.length === 0) return null;
    const index = Math.min(
        Math.max(activeSuggestionIndex.value, 0),
        prefixedSuggestions.value.length - 1,
    );
    return prefixedSuggestions.value[index] ?? null;
});

// Suggestions that add nothing beyond what's typed aren't worth a row.
const visibleSuggestions = computed(() =>
    inputFocused.value
        ? prefixedSuggestions.value
              .filter((s) => s.toLowerCase() !== trimmedItem.value.toLowerCase())
              .slice(0, 5)
        : [],
);

function isCompleting(id: number) {
    return completingIds.value.has(id);
}

function updateIdSet(id: number, value: boolean) {
    const source = completingIds;
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
    if (!value) return;
    // The row appears now; a failure rolls it back with a Retry toast (from the shared helper).
    createMutation.mutate({ client_id: newClientId(), name: value });
    newItem.value = "";
    activeSuggestionIndex.value = 0;
    await focusNewItemInput();
}

function onItemCheck(id: number, checked: boolean | "indeterminate") {
    // A row that hasn't been saved yet has no server id to complete.
    if (checked === true && !isTempGroceryId(id)) {
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
    updateIdSet(id, false);
    unhideItem(id);
}

async function completeItem(id: number) {
    if (isCompleting(id)) return;
    updateIdSet(id, true);
    const persistPromise = completeMutation.mutateAsync({ id });
    await new Promise((resolve) => setTimeout(resolve, COMPLETE_DELAY_MS));
    try {
        await persistPromise;
    } catch {
        // The helper already rolled the row back and toasted with Retry.
        clearItemState(id);
        return;
    }
    hideItem(id);
    updateIdSet(id, false);
}

function acceptSuggestion() {
    if (!activeSuggestion.value) return;
    newItem.value = activeSuggestion.value;
}

function onInputFocus() {
    inputFocused.value = true;
}

function onInputBlur() {
    // Delay so a tap on a suggestion lands before the list goes away.
    window.setTimeout(() => {
        inputFocused.value = false;
        activeSuggestionIndex.value = 0;
    }, 120);
}

function onInputKeydown(event: KeyboardEvent) {
    const hasInput = trimmedItem.value.length > 0;
    const count = prefixedSuggestions.value.length;
    const hasSuggestion = hasInput && activeSuggestion.value;

    if (event.key === "Tab" && hasSuggestion) {
        event.preventDefault();
        acceptSuggestion();
        return;
    }
    if (event.key === "ArrowRight" && hasSuggestion && event.target === newItemInput.value) {
        const input = newItemInput.value;
        if (input && input.selectionStart === input.value.length && input.selectionEnd === input.value.length) {
            event.preventDefault();
            acceptSuggestion();
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
        void addItem(activeSuggestion.value ?? undefined);
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
    <div class="mx-auto flex w-full max-w-3xl min-w-0 flex-col gap-4">
        <PageHeader title="Grocery list">
            <template #actions>
                <Button @click="openAddListDialog">
                    <template #icon-left><ClipboardList aria-hidden="true" class="size-4" /></template>
                    Paste list
                </Button>
            </template>
        </PageHeader>

        <Card padding="none">
            <ErrorState
                v-if="isError"
                :error="error"
                title="Couldn't load the grocery list"
                @retry="refetch()"
            />
            <SkeletonList v-else-if="showSkeleton" :rows="5" :leading="false" />
            <template v-else-if="!isPending">
                <EmptyState
                    v-if="displayItems.length === 0"
                    :icon="ShoppingBasket"
                    title="Nothing on the list"
                    description="Add what you need below, or paste a whole list."
                />
                <List v-else>
                    <ListRow
                        v-for="item in displayItems"
                        :key="item.client_id ?? item.ID"
                        :pending="isCompleting(item.ID)"
                    >
                        <Checkbox
                            :model-value="isCompleting(item.ID)"
                            :label="`Complete ${item.name}`"
                            @update:model-value="onItemCheck(item.ID, $event)"
                        >
                            <span
                                class="transition-colors duration-700"
                                :class="isCompleting(item.ID) ? 'text-fg-subtle line-through' : ''"
                            >{{ item.name }}</span>
                        </Checkbox>
                    </ListRow>
                </List>
                <div class="flex flex-col gap-2 border-t border-line p-3 lg:p-4" :class="displayItems.length === 0 ? 'border-t-0 pt-0' : ''">
                    <div
                        ref="addWrap"
                        @keydown="onInputKeydown"
                        @focusin="onInputFocus"
                        @focusout="onInputBlur"
                    >
                        <Input
                            v-model="newItem"
                            autocomplete="off"
                            aria-label="Add grocery item"
                            placeholder="Add item…"
                            @input="activeSuggestionIndex = 0"
                        >

                        </Input>
                    </div>
                    <List v-if="visibleSuggestions.length" class="rounded-control border border-line">
                        <ListRow
                            v-for="(suggestion, i) in visibleSuggestions"
                            :key="suggestion"
                            :selected="suggestion === activeSuggestion"
                            @click="addItem(suggestion)"
                            @mouseenter="activeSuggestionIndex = i"
                        >
                            {{ suggestion }}
                        </ListRow>
                    </List>
                </div>
            </template>
        </Card>
    </div>
</template>
