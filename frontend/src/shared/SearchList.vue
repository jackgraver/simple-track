<script setup lang="ts">
import { searchListKeys } from "~/shared/searchListKeys";
import { Check, Plus } from "lucide-vue-next";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { Component } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import { useListNavigation } from "~/composables/useListNavigation";
import { EmptyState, ErrorState, Input, List, ListRow, SkeletonList, useDelayedFlag } from "~/shared/ui";

const props = defineProps<{
    route: string;
    onSelect: (item: any) => Promise<boolean>;
    onCreate?: (name: string) => Promise<boolean>;
    /** Verb for the create row, e.g. "Create and add". Defaults to "Create". */
    createLabel?: string;
    displayComponent?: Component;
    prefilter?: any[];
    autoFocus?: boolean;
    /** When set, TanStack Query uses this key (e.g. share cache with useWorkoutExercisesAllQuery). */
    queryKey?: readonly unknown[];
}>();

const search = ref("");
const searchInput = ref<{ $el: HTMLElement } | null>(null);
const hasAutoFocused = ref(false);
const itemsContainer = ref<HTMLElement | null>(null);
const isHovering = ref(false);
const isFocused = ref(false);

const excludedIds = computed(() => new Set(props.prefilter ?? []));

const requestPath = computed(() =>
    props.route.startsWith("/") ? props.route : `/${props.route}`,
);

const { data, isPending, isError, error, refetch } = useQuery({
    queryKey: computed(() =>
        props.queryKey !== undefined
            ? [...props.queryKey]
            : searchListKeys.path(requestPath.value),
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

const PICKED_CHECK_MS = 1500;
const pickedKey = ref<string | null>(null);
const pickedAnnouncement = ref("");
let pickedTimer: ReturnType<typeof setTimeout> | undefined;
let unmounted = false;
const rowKey = (item: any, index: number) => `${item.entry_kind ?? "food"}-${item.id ?? item.ID ?? index}`;
function flashPicked(key: string, name: string) {
    pickedKey.value = key;
    clearTimeout(pickedTimer);
    pickedTimer = setTimeout(() => (pickedKey.value = null), PICKED_CHECK_MS);
    // Clear first so a repeat pick of the same row is announced again.
    pickedAnnouncement.value = "";
    nextTick(() => (pickedAnnouncement.value = name));
    // The list is unfiltered once the search clears, so bring the picked row back into view.
    nextTick(() => itemsContainer.value?.querySelector(`[data-key="${key}"]`)?.scrollIntoView({ block: "nearest" }));
}
onBeforeUnmount(() => {
    unmounted = true;
    clearTimeout(pickedTimer);
});

const handleFunctionCall = async <T extends (arg: any) => Promise<boolean>>(
    fn?: T,
    args?: any,
    key?: string,
) => {
    if (!fn) return;
    const success = await fn(args);
    if (unmounted) return;
    if (success) {
        if (key) flashPicked(key, args?.name ?? "");
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
            rowKey(filteredList.value[activeIndex.value], activeIndex.value),
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
        const el = itemsContainer.value?.querySelector("[aria-current=true]");
        el?.scrollIntoView({ block: "nearest" });
    });
}

const showSkeleton = useDelayedFlag(isPending);

const focusSearchInput = () => {
    if (!props.autoFocus || hasAutoFocused.value || isPending.value || !searchInput.value) {
        return;
    }
    hasAutoFocused.value = true;
    nextTick(() => searchInput.value?.$el.querySelector("input")?.focus());
};

onMounted(focusSearchInput);
watch(isPending, (pending) => {
    if (!pending) focusSearchInput();
});
</script>

<template>
    <ErrorState v-if="isError" :error="error" title="Couldn't load the list" class="m-2" @retry="refetch()" />
    <div
        v-else
        class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden p-2"
        @mouseenter="isHovering = true"
        @mouseleave="isHovering = false"
    >
        <span class="sr-only" role="status" aria-live="polite">{{ pickedAnnouncement }}</span>
        <Input
            ref="searchInput"
            v-model="search"
            placeholder="Search"
            aria-label="Search"
            autocomplete="off"
            :disabled="isPending"
            @keydown="onInputKeydown"
            @focusin="isFocused = true"
            @focusout="isFocused = false"
        />
        <div ref="itemsContainer" class="min-h-0 overflow-auto max-lg:max-h-[min(45dvh,20rem)] lg:flex-1">
            <SkeletonList v-if="isPending && showSkeleton" :rows="4" :leading="false" />
            <List v-else-if="!isPending">
                <ListRow
                    v-for="(item, index) in filteredList"
                    :key="rowKey(item, index)"
                    :data-key="rowKey(item, index)"
                    :selected="activeIndex === index"
                    @click="handleFunctionCall(onSelect, item, rowKey(item, index))"
                >
                    <component :is="props.displayComponent" v-if="props.displayComponent" :item="item" />
                    <template v-else>{{ item.name || `#${item.id ?? item.ID}` }}</template>
                    <template #trailing>
                        <Check v-if="pickedKey === rowKey(item, index)" aria-hidden="true" class="size-4 text-success" />
                        <Plus v-else aria-hidden="true" class="size-4" />
                    </template>
                </ListRow>
                <ListRow
                    v-if="showCreate"
                    :selected="activeIndex === filteredList.length"
                    @click="handleFunctionCall(onCreate, search)"
                >
                    <template #leading><Plus aria-hidden="true" class="size-4" /></template>
                    {{ createLabel ?? "Create" }} "{{ search }}"
                </ListRow>
            </List>
            <EmptyState v-if="!isPending && !search && !list.length" title="Nothing here yet" />
            <EmptyState v-else-if="!isPending && search && !filteredList.length && !onCreate" :title="`${search} does not exist`" />
        </div>
    </div>
</template>
