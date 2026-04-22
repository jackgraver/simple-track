<script setup lang="ts">
import { ref } from "vue";
import type { CompositeFood, Food } from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import { useCreateCompositeFood } from "../queries/useMealMutations";
import { Trash2 } from "lucide-vue-next";

const props = defineProps<{
    onResolve: (cf: CompositeFood) => void;
    onCancel?: () => void;
}>();

const name = ref("");
const lines = ref<{ food_id: number; food: Food; amount: number }[]>([]);

const createMutation = useCreateCompositeFood();

const addFood = async (item: Food & { entry_kind?: string }): Promise<boolean> => {
    if (item.entry_kind === "composite") return false;
    const food = item as Food;
    const existing = lines.value.findIndex((l) => l.food_id === food.ID);
    if (existing !== -1) {
        lines.value = lines.value.map((l, i) =>
            i === existing ? { ...l, amount: l.amount + 1 } : l,
        );
        return true;
    }
    lines.value = [...lines.value, { food_id: food.ID, food, amount: 1 }];
    return true;
};

const removeLine = (i: number) => {
    lines.value = lines.value.filter((_, idx) => idx !== i);
};

const submit = async () => {
    const n = name.value.trim();
    if (!n || lines.value.length === 0) return;
    try {
        const res = await createMutation.mutateAsync({
            name: n,
            items: lines.value.map((l) => ({
                food_id: l.food_id,
                amount: l.amount,
            })),
        });
        if (res?.composite_food) props.onResolve(res.composite_food);
    } catch (e) {
        console.error(e);
    }
};
</script>

<template>
    <div class="composite-form">
        <div class="field">
            <label for="recname">Recipe name</label>
            <input id="recname" type="text" v-model="name" placeholder="e.g. Caesar Sauce" />
        </div>
        <p class="hint">Add foods below (same search as log meal).</p>
        <div class="picker">
            <SearchList
                route="diet/meals/food/all"
                :on-select="addFood"
                :display-component="FoodDisplay"
            />
        </div>
        <ul v-if="lines.length" class="lines">
            <li v-for="(line, i) in lines" :key="line.food_id" class="line">
                <span class="line-name">{{ line.food.name }}</span>
                <input
                    class="amt"
                    type="number"
                    min="0.01"
                    step="any"
                    v-model.number="line.amount"
                />
                <button type="button" class="rm" aria-label="Remove" @click="removeLine(i)">
                    <Trash2 :size="18" />
                </button>
            </li>
        </ul>
        <button type="button" class="submit" :disabled="!name.trim() || lines.length === 0" @click="submit">
            Save recipe
        </button>
    </div>
</template>

<style scoped>
.composite-form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    color: white;
    padding: 1rem 0.5rem;
    max-height: 70dvh;
}
.field {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}
.hint {
    margin: 0;
    font-size: 0.85rem;
    color: #aaa;
}
.picker {
    min-height: 8rem;
    max-height: 14rem;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}
.lines {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: 10rem;
    overflow-y: auto;
}
.line {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
}
.line-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
}
.amt {
    width: 4rem;
    background: #323232;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.25rem;
}
.rm {
    background: transparent;
    border: none;
    color: #ccc;
    cursor: pointer;
    padding: 0.25rem;
}
.rm:hover {
    color: #f87171;
}
.submit {
    margin-top: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    border: none;
    background: #3b82f6;
    color: white;
    cursor: pointer;
    font-weight: 600;
}
.submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
input[type="text"] {
    background: #323232;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.5rem;
}
label {
    font-size: 0.85rem;
    color: #ccc;
}
</style>
