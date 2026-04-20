<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from "vue";
import type { Food } from "~/types/diet";
import {
    searchFoods,
    type FdcFoodSummary,
} from "~/api/diet/food";
import { useCreateFood } from "../queries/useFoodMutations";

const props = defineProps<{
    onResolve: (food: Food) => void;
    onCancel?: () => void;
}>();

const query = ref("");
const results = ref<FdcFoodSummary[]>([]);
const loading = ref(false);
const error = ref("");
/** Trimmed query of the last successful search (for empty-state messaging). */
const lastSearchQuery = ref("");
const selectingFdcId = ref<number | null>(null);

const createFoodMutation = useCreateFood();

let abortController: AbortController | null = null;

watch(query, () => {
    results.value = [];
    error.value = "";
    lastSearchQuery.value = "";
});

function mapSummaryToFood(row: FdcFoodSummary): Food {
    const name = row.brand?.trim()
        ? `${row.name.trim()} (${row.brand.trim()})`
        : row.name.trim();
    const unit = (row.serving_size_unit || "g").toLowerCase();
    return {
        ID: 0,
        created_at: "",
        updated_at: "",
        name,
        serving_type: unit === "g" ? "g" : unit,
        serving_amount: row.serving_size,
        calories: row.calories,
        protein: row.protein_g,
        fiber: row.fiber_g,
        carbs: row.carbs_g,
    };
}

function submitSearch() {
    void runSearch(query.value);
}

async function runSearch(raw: string) {
    const q = raw.trim();
    if (q.length < 2) {
        results.value = [];
        lastSearchQuery.value = "";
        error.value = "Type at least 2 characters, then tap Search.";
        return;
    }

    abortController?.abort();
    abortController = new AbortController();
    const signal = abortController.signal;

    loading.value = true;
    error.value = "";
    try {
        const data = await searchFoods(q, signal);
        results.value = data.foods ?? [];
        lastSearchQuery.value = q;
    } catch (e: unknown) {
        const err = e as { code?: string; name?: string; message?: string };
        if (
            err?.code === "ERR_CANCELED" ||
            err?.name === "CanceledError" ||
            err?.name === "AbortError"
        ) {
            return;
        }
        results.value = [];
        lastSearchQuery.value = "";
        error.value = err?.message || "Search failed";
    } finally {
        loading.value = false;
    }
}

onBeforeUnmount(() => {
    abortController?.abort();
});

async function selectRow(row: FdcFoodSummary) {
    selectingFdcId.value = row.fdc_id;
    try {
        const payload = mapSummaryToFood(row);
        const response = await createFoodMutation.mutateAsync(payload);
        if (response?.food) {
            props.onResolve(response.food);
        }
    } catch {
        /* useCreateFood already toasts */
    } finally {
        selectingFdcId.value = null;
    }
}

function macroLine(row: FdcFoodSummary): string {
    const parts = [
        `${Math.round(row.calories)} kcal`,
        `${formatG(row.protein_g)} P`,
        `${formatG(row.carbs_g)} C`,
        `${formatG(row.fiber_g)} fiber`,
    ];
    return parts.join(" · ");
}

function formatG(n: number): string {
    if (!Number.isFinite(n)) return "0";
    return n >= 10 ? String(Math.round(n)) : n.toFixed(1);
}

function servingLine(row: FdcFoodSummary): string {
    const u = row.serving_size_unit || "g";
    const n = row.serving_size;
    if (!Number.isFinite(n) || n <= 0) return "per 100 g";
    const num = n === Math.floor(n) ? String(n) : n.toFixed(1);
    return `per ${num} ${u}`;
}
</script>

<template>
    <div class="search-food">
        <div class="search-form">
            <label for="food-search-q">Search foods</label>
            <input
                id="food-search-q"
                v-model="query"
                type="text"
                inputmode="search"
                autocomplete="off"
                placeholder="e.g. raspberries"
                class="search-input"
            />
            <button
                type="button"
                class="search-btn-physical"
                :disabled="loading"
                @click.prevent.stop="submitSearch"
            >
                Search foods
            </button>
        </div>
        <p v-if="loading" class="hint">Searching…</p>
        <p v-else-if="error" class="err">{{ error }}</p>
        <ul v-if="results.length" class="list" role="listbox">
            <li v-for="row in results" :key="row.fdc_id">
                <button
                    type="button"
                    class="row-btn"
                    :disabled="selectingFdcId !== null"
                    @click="selectRow(row)"
                >
                    <span class="row-title">{{ row.name }}</span>
                    <span v-if="row.brand" class="row-brand">{{ row.brand }}</span>
                    <span class="row-serving">{{ servingLine(row) }}</span>
                    <span class="row-macros">{{ macroLine(row) }}</span>
                </button>
            </li>
        </ul>
        <p
            v-else-if="
                lastSearchQuery &&
                lastSearchQuery === query.trim() &&
                !loading &&
                !error
            "
            class="hint"
        >
            No results
        </p>
    </div>
</template>

<style scoped>
.search-food {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    color: white;
    min-width: min(100%, 22rem);
}

.search-form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.search-input {
    width: 100%;
    background-color: rgb(50, 50, 50);
    color: white;
    border: 1px solid #4a4a4a;
    border-radius: 6px;
    padding: 0.65rem 0.75rem;
    box-sizing: border-box;
    font: inherit;
}

.search-input:focus {
    outline: none;
    border-color: #6a9fd4;
}

.search-btn-physical {
    display: block;
    width: 100%;
    margin: 0;
    padding: 0.75rem 1rem;
    min-height: 2.75rem;
    border: 2px solid #3d7ab8;
    border-radius: 8px;
    background: linear-gradient(180deg, #3a7ec4 0%, #2d6cb0 100%);
    color: #fff;
    font: inherit;
    font-weight: 600;
    font-size: 1rem;
    cursor: pointer;
    box-shadow:
        0 1px 0 rgba(255, 255, 255, 0.12) inset,
        0 2px 4px rgba(0, 0, 0, 0.35);
    -webkit-tap-highlight-color: transparent;
    appearance: none;
}

.search-btn-physical:hover:not(:disabled) {
    background: linear-gradient(180deg, #4a8ed4 0%, #3d7cc0 100%);
}

.search-btn-physical:active:not(:disabled) {
    transform: translateY(1px);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.4) inset;
}

.search-btn-physical:disabled {
    opacity: 0.55;
    cursor: not-allowed;
    box-shadow: none;
}

label {
    font-size: 0.9rem;
    color: #ccc;
}

.hint {
    margin: 0;
    font-size: 0.85rem;
    color: #999;
}

.err {
    margin: 0;
    font-size: 0.85rem;
    color: #f66;
}

.list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: 16rem;
    overflow-y: auto;
}

.row-btn {
    width: 100%;
    text-align: left;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 0.15rem;
    padding: 0.55rem 0.65rem;
    border: 1px solid #3d3d3d;
    border-radius: 6px;
    background: rgb(38, 38, 38);
    color: #eee;
    cursor: pointer;
    font: inherit;
}

.row-btn:hover:not(:disabled) {
    background: rgb(48, 48, 48);
    border-color: #555;
}

.row-btn:disabled {
    opacity: 0.6;
    cursor: wait;
}

.row-title {
    font-weight: 600;
    font-size: 0.95rem;
}

.row-brand {
    font-size: 0.8rem;
    color: #aaa;
}

.row-serving {
    font-size: 0.75rem;
    color: #888;
}

.row-macros {
    font-size: 0.8rem;
    color: #bbb;
    font-variant-numeric: tabular-nums;
}
</style>
