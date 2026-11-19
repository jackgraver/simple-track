<script lang="ts">
export type ComboboxOption = { label: string; value: string | number };
</script>

<script setup lang="ts">
// Search-and-pick: type to filter, arrow keys + Enter (or tap) to choose. Reka Combobox gives the
// listbox ARIA, keyboard nav, portal and positioning. Action-style: choosing emits `select` and
// clears the input, so it does not hold a value. Server-side search and paging are opt-in through
// `search` (debounced), `has-more`/`loading` and `load-more`.
import { ref, computed, watch, onBeforeUnmount } from "vue";
import {
    ComboboxAnchor,
    ComboboxContent,
    ComboboxEmpty,
    ComboboxInput,
    ComboboxItem,
    ComboboxPortal,
    ComboboxRoot,
    ComboboxViewport,
} from "reka-ui";
import { Search } from "lucide-vue-next";
import { useField } from "./fieldContext";
import Spinner from "./Spinner.vue";

const props = withDefaults(
    defineProps<{
        options: ComboboxOption[];
        placeholder?: string;
        hasMore?: boolean;
        loading?: boolean;
        disabled?: boolean;
        ariaLabel?: string;
        emptyText?: string;
    }>(),
    { placeholder: "Search…", hasMore: false, loading: false, emptyText: "No matches" },
);
const emit = defineEmits<{
    select: [option: ComboboxOption];
    search: [query: string];
    "load-more": [];
}>();

const field = useField();
const term = ref("");
const open = ref(false);

const filtered = computed(() => {
    const q = term.value.toLowerCase().trim();
    return q ? props.options.filter((o) => o.label.toLowerCase().includes(q)) : props.options;
});

function onPick(value: unknown) {
    const option = props.options.find((o) => o.value === value);
    if (!option) return;
    emit("select", option);
    term.value = "";
    open.value = false;
}

function onScroll(e: Event) {
    const el = e.target as HTMLElement;
    if (props.loading || !props.hasMore) return;
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 20) emit("load-more");
}

let timer: ReturnType<typeof setTimeout> | undefined;
watch(term, (q) => {
    clearTimeout(timer);
    timer = setTimeout(() => emit("search", q), 300);
});
onBeforeUnmount(() => clearTimeout(timer));

// `!` classes beat the global `input {}` rule in App.vue (phase 6 removes it), same as Input.vue.
const control =
    "m-0! min-w-0 flex-1 rounded-none! border-0! bg-transparent! p-0! text-base text-fg! outline-none! placeholder:text-fg-subtle lg:text-sm";
const itemClass =
    "flex min-h-touch cursor-pointer items-center rounded-control px-3 text-base text-fg outline-none select-none data-[highlighted]:bg-surface-3 lg:min-h-row lg:text-sm";
</script>

<template>
    <ComboboxRoot
        v-model:open="open"
        :model-value="null"
        ignore-filter
        open-on-focus
        open-on-click
        :disabled="disabled"
        @update:model-value="onPick"
    >
        <ComboboxAnchor
            class="flex min-h-touch items-center gap-2 rounded-control border border-line-strong bg-surface-2 px-3 transition-colors duration-150 focus-within:border-accent focus-within:ring-1 focus-within:ring-accent lg:min-h-row"
            :class="disabled ? 'cursor-not-allowed opacity-50' : ''"
        >
            <Search aria-hidden="true" class="size-4 shrink-0 text-fg-muted" />
            <ComboboxInput
                v-model="term"
                :id="field?.id"
                :placeholder="placeholder"
                :aria-label="ariaLabel ?? placeholder"
                autocomplete="off"
                :class="control"
            />
        </ComboboxAnchor>
        <ComboboxPortal>
            <ComboboxContent
                position="popper"
                :side-offset="6"
                :collision-padding="8"
                class="z-popover w-(--reka-combobox-trigger-width) rounded-card border border-line bg-surface shadow-raised"
            >
                <ComboboxViewport class="max-h-60 overflow-y-auto p-1" @scroll="onScroll">
                    <ComboboxEmpty v-if="!filtered.length && !loading" class="px-3 py-3 text-sm text-fg-muted">
                        {{ emptyText }}
                    </ComboboxEmpty>
                    <ComboboxItem v-for="o in filtered" :key="o.value" :value="o.value" :class="itemClass">
                        <span class="min-w-0 flex-1 truncate">{{ o.label }}</span>
                    </ComboboxItem>
                    <div v-if="loading" class="flex justify-center py-3"><Spinner size="sm" /></div>
                    <p v-else-if="!hasMore && filtered.length" class="m-0 px-3 py-2 text-center text-xs text-fg-subtle">
                        No more results
                    </p>
                </ComboboxViewport>
            </ComboboxContent>
        </ComboboxPortal>
    </ComboboxRoot>
</template>
