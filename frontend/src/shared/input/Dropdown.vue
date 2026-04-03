<script lang="ts">
export type DropdownOption = {
    label: string;
    value: string | number;
};
</script>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { Loader } from "lucide-vue-next";

const props = withDefaults(
    defineProps<{
        options: DropdownOption[];
        placeholder?: string;
        onSelect: (option: DropdownOption) => void;
        hasMore?: boolean;
        loading?: boolean;
    }>(),
    {
        placeholder: "Search...",
        hasMore: false,
        loading: false,
    },
);

const emit = defineEmits<{
    (e: "load-more"): void;
    (e: "search", query: string): void;
}>();

const searchQuery = ref("");
const showDropdown = ref(false);
const listRef = ref<HTMLElement | null>(null);

const filteredOptions = computed(() => {
    const query = searchQuery.value.toLowerCase().trim();
    if (!query) return props.options;
    return props.options.filter((opt) =>
        opt.label.toLowerCase().includes(query),
    );
});

const selectOption = (option: DropdownOption) => {
    props.onSelect(option);
    searchQuery.value = "";
    showDropdown.value = false;
};

const handleBlur = () => {
    setTimeout(() => {
        showDropdown.value = false;
    }, 200);
};

const handleScroll = () => {
    const el = listRef.value;
    if (!el || props.loading || !props.hasMore) return;
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 20) {
        emit("load-more");
    }
};

let searchTimeout: ReturnType<typeof setTimeout>;
watch(searchQuery, (query) => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        emit("search", query);
    }, 300);
});
</script>

<template>
    <div class="relative w-full">
        <input
            v-model="searchQuery"
            @focus="showDropdown = true"
            @blur="handleBlur"
            type="text"
            :placeholder="placeholder"
            class="w-full p-3 border border-[#383838] rounded-[5px] bg-[#1b1b1b] text-inherit text-base box-border focus:outline-none focus:border-[#646464] focus:bg-[#232323]"
        />
        <ul
            v-if="showDropdown && (filteredOptions.length > 0 || loading)"
            ref="listRef"
            @scroll="handleScroll"
            class="absolute top-full left-0 right-0 mt-1 list-none p-0 bg-[#1b1b1b] border border-[#383838] rounded-[5px] max-h-[200px] overflow-y-auto z-1000"
        >
            <li
                v-for="option in filteredOptions"
                :key="option.value"
                @mousedown.prevent="selectOption(option)"
                class="p-3 cursor-pointer transition-colors duration-200 hover:bg-[#282828]"
            >
                {{ option.label }}
            </li>
            <li v-if="loading" class="p-3 flex justify-center items-center">
                <Loader class="animate-spin" :size="18" />
            </li>
            <li
                v-else-if="!hasMore && filteredOptions.length > 0"
                class="p-2 text-center text-xs text-neutral-500"
            >
                No more results
            </li>
        </ul>
    </div>
</template>
