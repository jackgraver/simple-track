<script lang="ts">
export type DropdownOption = {
    label: string;
    value: string | number;
};
</script>

<script setup lang="ts">
// Compatibility wrapper: same props and emits as before, now the shared/ui Combobox. New code
// should use Combobox directly (its `select` emit instead of the `onSelect` prop).
import { Combobox } from "~/shared/ui";

const props = withDefaults(
    defineProps<{
        options: DropdownOption[];
        placeholder?: string;
        onSelect: (option: DropdownOption) => void;
        hasMore?: boolean;
        loading?: boolean;
    }>(),
    { placeholder: "Search...", hasMore: false, loading: false },
);

const emit = defineEmits<{
    (e: "load-more"): void;
    (e: "search", query: string): void;
}>();
</script>

<template>
    <Combobox
        :options="options"
        :placeholder="placeholder"
        :has-more="hasMore"
        :loading="loading"
        @select="props.onSelect"
        @search="emit('search', $event)"
        @load-more="emit('load-more')"
    />
</template>
