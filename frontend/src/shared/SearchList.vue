<script setup lang="ts">
import { Plus, Loader } from "lucide-vue-next";
import { computed, ref } from "vue";
import type { Component } from "vue";
import { useAPIGet } from "~/composables/useApiFetch";

const props = defineProps<{
    route: string;
    onSelect: (item: any) => Promise<boolean>;
    onCreate?: (name: string) => Promise<boolean>;
    displayComponent?: Component;
    prefilter?: any[];
}>();

const search = ref("");

const query = props.prefilter
    ? "?" + props.prefilter.map((id) => `exclude=${id}`).join("&")
    : "";
const { data, pending, error } = useAPIGet<any[]>(props.route + query);

const list = computed(() => {
    const value = data.value;
    if (!value) return [];

    if (Array.isArray(value)) return value;

    const firstArray = Object.values(value).find((v) => Array.isArray(v));
    return firstArray ?? [];
});

const filteredList = computed(() => {
    const term = search.value.trim().toLowerCase();
    return term
        ? list.value.filter((f) => f.name.toLowerCase().includes(term))
        : list.value;
});

const handleFunctionCall = async <T extends (arg: any) => Promise<boolean>>(
    fn?: T,
    args?: any,
) => {
    if (!fn) return;
    const success = await fn(args);
    if (success) search.value = "";
};

const refresh = () => {
    console.log("refresh");
};
</script>

<template>
    <div v-if="error" class="search-container">
        <span class="error-message">Failed to Load Saved Meals</span>
        <button @click="refresh">Try again</button>
    </div>
    <div v-else class="search-container">
        <div class="search-input-wrapper">
            <input
                type="text"
                v-model="search"
                placeholder="Search"
                :disabled="pending"
            />
        </div>
        <div class="items-container">
            <template v-if="pending">
                <Loader v-if="pending" class="spinner" :size="32" />
            </template>
            <template v-else-if="filteredList?.length">
                <button
                    v-for="(item, index) in filteredList"
                    :key="item.id ?? item.ID ?? index"
                    @click="handleFunctionCall(onSelect, item)"
                    class="item"
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
            </template>
            <div v-else-if="!pending && search" class="item empty-option">
                <template v-if="onCreate">
                    <button
                        type="button"
                        class="create-button"
                        @click="handleFunctionCall(onCreate, search)"
                    >
                        <Plus :size="18" />
                        <span>Create "{{ search }}"</span>
                    </button>
                </template>
                <template v-else-if="pending">
                    <Loader class="spinner" :size="18" />
                </template>
                <template v-else>
                    <span class="no-hover-empty-option"
                        >{{ search }} does not exist</span
                    >
                </template>
            </div>
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

.item:hover:not(:has(.no-hover-empty-option)) {
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

.create-button {
    display: flex;
    align-items: center;
    gap: 6px;
    justify-content: center;
    padding: 0.6rem 0.8rem;
    cursor: pointer;
    background-color: transparent;
    box-shadow: none;
}
.create-button:hover {
    background-color: transparent;
}
</style>
