<script setup lang="ts">
import { Plus } from "lucide-vue-next";

const props = defineProps<{
    route: string;
    onSelect: (item: any) => Promise<boolean>;
    onCreate?: (name: string) => Promise<boolean>;
    displayComponent?: Component;
    prefilter?: any[];
}>();

const search = ref("");

const query = props.prefilter?.map((id) => `exclude=${id}`).join("&");
const { data, pending, error } = useAPIGet<any[]>(props.route + "?" + query);

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
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="search-container">
        <input type="text" v-model="search" placeholder="Search" />
        <div class="items-container">
            <template v-if="filteredList?.length">
                <button
                    v-for="(item, index) in filteredList"
                    :key="item.id ?? item.ID ?? index"
                    @click="
                        props.onSelect(item).then((res) => {
                            if (res) search = '';
                        })
                    "
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
            <div v-else class="item empty-option">
                <template v-if="onCreate">
                    <button
                        class="create-button"
                        @click="
                            onCreate(search).then((res) => {
                                if (res) search = '';
                            })
                        "
                    >
                        <Plus :size="18" />
                        <span>Create “{{ search }}”</span>
                    </button>
                </template>
                <template v-else>
                    <span class="no-hover-empty-option"
                        >“{{ search }}” does not exist</span
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
    gap: 1rem;
    width: 100%;
}

.items-container {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
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
    background-color: none;
}
</style>
