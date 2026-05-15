<script setup lang="ts">
import { Plus, Loader } from "lucide-vue-next";
import { computed, ref, nextTick } from "vue";
import type { Component } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import { useListNavigation } from "~/composables/useListNavigation";

const props = defineProps<{
    route: string;
    onSelect: (item: any) => Promise<boolean>;
    onCreate?: (name: string) => Promise<boolean>;
    displayComponent?: Component;
    prefilter?: any[];
    /** When set, TanStack Query uses this key (e.g. share cache with useWorkoutExercisesAllQuery). */
    queryKey?: readonly unknown[];
}>();

const search = ref("");
const itemsContainer = ref<HTMLElement | null>(null);
const isHovering = ref(false);
const isFocused = ref(false);

const excludedIds = computed(() => new Set(props.prefilter ?? []));

const querySuffix = computed(() =>
    props.prefilter?.length
        ? "?" + props.prefilter.map((id) => `exclude=${id}`).join("&")
        : "",
);

const requestPath = computed(() => {
    const base = props.route.startsWith("/") ? props.route : `/${props.route}`;
    return `${base}${querySuffix.value}`;
});

const { data, isPending, error, refetch } = useQuery({
    queryKey: computed(() =>
        props.queryKey !== undefined
            ? [...props.queryKey]
            : (["searchList", requestPath.value] as const),
    ),
    queryFn: async () => {
        const res = await apiClient.get<unknown>(requestPath.value);
        return res.data;
    },
    enabled: computed(() => !!props.route),
});

const list = computed(() => {
    const value = data.value;
    if (!value) return [];

    if (Array.isArray(value)) {
        return value.filter(
            (item) => !excludedIds.value.has(item.id ?? item.ID),
        );
    }

    const obj = value as Record<string, unknown>;
    const hasFoodsKey = Object.prototype.hasOwnProperty.call(obj, "foods");
    const hasCompositeKey = Object.prototype.hasOwnProperty.call(
        obj,
        "composite_foods",
    );
    if (hasFoodsKey || hasCompositeKey) {
        const foods = Array.isArray(obj.foods) ? obj.foods : [];
        const composites = Array.isArray(obj.composite_foods)
            ? obj.composite_foods
            : [];
        const withKindFood = foods.map((f: Record<string, unknown>) => ({
            ...f,
            entry_kind: f.entry_kind ?? "food",
        }));
        const merged = [...withKindFood, ...composites];
        return merged.filter(
            (item: { id?: number; ID?: number }) =>
                !excludedIds.value.has(item.id ?? item.ID),
        );
    }

    const firstArray = Object.values(value as object).find((v) =>
        Array.isArray(v),
    );
    return (firstArray ?? []).filter(
        (item) => !excludedIds.value.has(item.id ?? item.ID),
    );
});

const filteredList = computed(() => {
    const term = search.value.trim().toLowerCase();
    return term
        ? list.value.filter((f) => f.name.toLowerCase().includes(term))
        : list.value;
});

const showCreate = computed(
    () => !isPending.value && search.value.trim().length > 0 && props.onCreate,
);

const navigableCount = computed(
    () => filteredList.value.length + (showCreate.value ? 1 : 0),
);

const globalNavActive = computed(() => isHovering.value && !isFocused.value);
const { activeIndex, handleKey, setEnterHandler, reset } = useListNavigation(
    navigableCount,
    globalNavActive,
);

const handleFunctionCall = async <T extends (arg: any) => Promise<boolean>>(
    fn?: T,
    args?: any,
) => {
    if (!fn) return;
    const success = await fn(args);
    if (success) {
        search.value = "";
        reset();
    }
};

function selectActive() {
    if (activeIndex.value === -1) return;
    const createIdx = filteredList.value.length;
    if (activeIndex.value < createIdx) {
        handleFunctionCall(
            props.onSelect,
            filteredList.value[activeIndex.value],
        );
    } else if (showCreate.value) {
        handleFunctionCall(props.onCreate, search.value);
    }
}

setEnterHandler(selectActive);

function onInputKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
        e.preventDefault();
        selectActive();
        return;
    }
    handleKey(e);
    nextTick(() => {
        const el = itemsContainer.value?.querySelector(".item.active");
        el?.scrollIntoView({ block: "nearest" });
    });
}

const refresh = () => {
    refetch();
};
</script>

<template>
    <div v-if="error" class="search-container">
        <span class="error-message">Failed to Load Saved Meals</span>
        <button @click="refresh">Try again</button>
    </div>
    <div
        v-else
        class="search-container"
        @mouseenter="isHovering = true"
        @mouseleave="isHovering = false"
    >
        <div class="search-input-wrapper">
            <input
                type="text"
                v-model="search"
                placeholder="Search"
                :disabled="isPending"
                @keydown="onInputKeydown"
                @focus="isFocused = true"
                @blur="isFocused = false"
            />
        </div>
        <div
            class="items-container min-h-0 max-md:max-h-[min(45dvh,20rem)] md:flex-1"
            ref="itemsContainer"
        >
            <template v-if="isPending">
                <Loader class="spinner" :size="32" />
            </template>
            <template v-else>
                <button
                    v-for="(item, index) in filteredList"
                    :key="`${item.entry_kind ?? 'food'}-${item.id ?? item.ID ?? index}`"
                    @click="handleFunctionCall(onSelect, item)"
                    class="item"
                    :class="{ active: activeIndex === index }"
                    role="option"
                >
                    <component
                        v-if="props.displayComponent"
                        :is="props.displayComponent"
                        :item="item"
                    />
                    <template v-else>
                        {{ item.name || `#${item.id ?? item.ID}` }}
                    </template>
                    <Plus :size="18" />
                </button>
                <div
                    v-if="showCreate"
                    class="item empty-option"
                    :class="{ active: activeIndex === filteredList.length }"
                    @click="handleFunctionCall(onCreate, search)"
                >
                    <Plus :size="18" />
                    <span>Create "{{ search }}"</span>
                </div>
                <div
                    v-if="
                        !isPending &&
                        search &&
                        !filteredList.length &&
                        !onCreate
                    "
                    class="item empty-option"
                >
                    <span class="no-hover-empty-option"
                        >{{ search }} does not exist</span
                    >
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped>
.search-container {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-height: 0;
    gap: 1rem;
    padding: 0.5rem;
    overflow: hidden;
}

.search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
}

.search-input-wrapper input {
    width: 100%;
}

.search-input-wrapper input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.spinner {
    position: absolute;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}

.error-message {
    color: #ff6b6b;
    padding: 1rem;
    text-align: center;
}

.loading-state {
    justify-content: center;
    opacity: 0.6;
    cursor: default;
}

.items-container {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
    overflow: auto;
}

.item {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border: 1px solid rgb(82, 82, 82);
    border-radius: 5px;
    background-color: rgb(48, 48, 48);
    cursor: pointer;
    transition: background-color 0.2s ease-in-out;
}

.item:hover:not(:has(.no-hover-empty-option)),
.item.active:not(:has(.no-hover-empty-option)) {
    background-color: rgb(82, 82, 82);
}

.empty-option {
    display: flex;
    align-items: center;
    gap: 6px;
    justify-content: center;
    padding: 0.6rem 0.8rem;
    cursor: pointer;
}

.no-hover-empty-option {
    cursor: default;
}
</style>
