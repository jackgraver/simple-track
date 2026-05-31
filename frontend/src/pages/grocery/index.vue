<script setup lang="ts">
import { computed, ref } from "vue";
import {
    useCompleteGroceryItem,
    useCreateGroceryItem,
    useDeleteGroceryItem,
    useGroceryItems,
    useGrocerySuggestions,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";

const newItem = ref("");
const completingIds = ref<Set<number>>(new Set());

const { data, isError, isPending, error } = useGroceryItems();
const suggestionsQuery = useGrocerySuggestions(newItem);
const createMutation = useCreateGroceryItem();
const completeMutation = useCompleteGroceryItem();
const deleteMutation = useDeleteGroceryItem();

const items = computed(() =>
    (data.value ?? []).filter((item) => !completingIds.value.has(item.ID)),
);
const suggestions = computed(() => suggestionsQuery.data.value ?? []);
const trimmedItem = computed(() => newItem.value.trim());
const canAdd = computed(
    () => trimmedItem.value.length > 0 && !createMutation.isPending.value,
);

function setCompleting(id: number, value: boolean) {
    const next = new Set(completingIds.value);
    if (value) {
        next.add(id);
    } else {
        next.delete(id);
    }
    completingIds.value = next;
}

async function addItem() {
    if (!canAdd.value) return;
    try {
        await createMutation.mutateAsync({ name: trimmedItem.value });
        newItem.value = "";
    } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to add item";
        toast.push(message, "error");
    }
}

async function completeItem(id: number) {
    setCompleting(id, true);
    try {
        await completeMutation.mutateAsync(id);
    } catch (err) {
        setCompleting(id, false);
        const message =
            err instanceof Error ? err.message : "Failed to complete item";
        toast.push(message, "error");
    }
}

async function removeItem(id: number) {
    try {
        await deleteMutation.mutateAsync(id);
    } catch (err) {
        const message =
            err instanceof Error ? err.message : "Failed to remove item";
        toast.push(message, "error");
    }
}
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-6">
        <div
            class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
        >
            <h1 class="m-0 text-lg font-semibold text-textPrimary">
                Grocery List
            </h1>
        </div>
        <form
            class="flex flex-col gap-3 rounded-md border border-(--color-border) bg-firstBg p-4 sm:flex-row"
            @submit.prevent="addItem"
        >
            <label class="flex min-w-0 flex-1 flex-col gap-1 text-xs text-textSecondary"
                >Item
                <input
                    v-model="newItem"
                    type="text"
                    list="grocery-suggestions"
                    autocomplete="off"
                    placeholder="e.g. ground beef"
                    class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                />
            </label>
            <datalist id="grocery-suggestions">
                <option
                    v-for="suggestion in suggestions"
                    :key="suggestion"
                    :value="suggestion"
                />
            </datalist>
            <button
                type="submit"
                class="self-end rounded-md bg-secondBg px-4 py-2 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg disabled:opacity-50"
                :disabled="!canAdd"
            >
                Add
            </button>
        </form>
        <div v-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load grocery list" }}
        </div>
        <div v-else-if="isPending" class="text-sm text-textSecondary">
            Loading…
        </div>
        <section
            v-else
            class="flex flex-col gap-2 rounded-md border border-(--color-border) bg-firstBg p-4"
        >
            <p v-if="items.length === 0" class="m-0 text-sm text-textSecondary">
                Nothing on the list.
            </p>
            <div
                v-for="item in items"
                :key="item.ID"
                class="flex items-center justify-between gap-3 rounded-md bg-secondBg px-3 py-2"
            >
                <label class="flex min-w-0 flex-1 items-center gap-3 text-sm text-textPrimary">
                    <input
                        type="checkbox"
                        class="m-0 size-4"
                        @change="completeItem(item.ID)"
                    />
                    <span class="truncate">{{ item.name }}</span>
                </label>
                <button
                    type="button"
                    class="p-0! text-xs text-textSecondary transition-colors hover:text-textPrimary"
                    @click="removeItem(item.ID)"
                >
                    Remove
                </button>
            </div>
        </section>
    </div>
</template>
