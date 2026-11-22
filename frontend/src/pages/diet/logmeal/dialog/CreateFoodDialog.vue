<script setup lang="ts">
import { ref } from "vue";
import type { Food } from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import { useCreateFood } from "~/api/diet/queries";
import { ChevronDown, ChevronUp } from "lucide-vue-next";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { Badge, Button, Field, IconButton, Input, Select } from "~/shared/ui";

const props = defineProps<{
    foodName?: string;
    onResolve: (food: Food) => void;
    onCancel?: () => void;
}>();

const meal = ref<Food>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: props.foodName ?? "",
    calories: 0,
    protein: 0,
    fiber: 0,
    carbs: 0,
    fat: 0,
    serving_type: "",
    serving_amount: 0,
});

const relatedOpen = ref(false);
const relatedFood = ref<Food | null>(null);

const createFoodMutation = useCreateFood();
const servingOptions = [
    { label: "Unit", value: "Unit" },
    { label: "Grams", value: "Grams" },
];

function clearRelated() {
    relatedFood.value = null;
    relatedOpen.value = false;
}

async function onPickRelated(
    row: Food & { entry_kind?: string },
): Promise<boolean> {
    if (row.entry_kind === "composite") return false;
    relatedFood.value = row as Food;
    relatedOpen.value = false;
    return true;
}

const createFood = async () => {
    if (meal.value.serving_type === "Grams") meal.value.serving_type = "g";
    if (meal.value.serving_type === "Unit") meal.value.serving_type = "";

    try {
        const response = await createFoodMutation.mutateAsync({
            food: meal.value,
            relatedFoodId: relatedFood.value?.ID,
        });
        if (response?.food) {
            props.onResolve(response.food);
        }
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Could not create food"), "error");
    }
};
</script>

<template>
    <form class="flex flex-col gap-4 text-left" @submit.prevent="createFood">
        <Field label="Food name" required>
            <Input v-model="meal.name" required autocomplete="off" />
        </Field>
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-5">
            <Field label="Calories">
                <Input v-model="meal.calories" type="number" inputmode="decimal" />
            </Field>
            <Field label="Protein">
                <Input v-model="meal.protein" type="number" inputmode="decimal" />
            </Field>
            <Field label="Carbs">
                <Input v-model="meal.carbs" type="number" inputmode="decimal" />
            </Field>
            <Field label="Fat">
                <Input v-model="meal.fat" type="number" inputmode="decimal" />
            </Field>
            <Field label="Fiber">
                <Input v-model="meal.fiber" type="number" inputmode="decimal" />
            </Field>
        </div>
        <Field label="Serving type">
            <Select
                v-model="meal.serving_type"
                :options="servingOptions"
                placeholder="Select serving type"
            />
        </Field>
        <Field
            v-if="meal.serving_type === 'Unit' || meal.serving_type === 'Grams'"
            label="Serving amount"
        >
            <Input v-model="meal.serving_amount" type="number" inputmode="decimal" />
        </Field>
        <div class="flex flex-col gap-2">
            <div class="flex items-center justify-between gap-2">
                <span class="min-w-0 flex-1 text-sm font-medium text-fg-muted"
                    >Related to existing food (optional)</span
                >
                <IconButton
                    :label="relatedOpen ? 'Hide related food search' : 'Search related food'"
                    variant="secondary"
                    size="sm"
                    :aria-expanded="relatedOpen"
                    aria-controls="related-search-panel"
                    @click="relatedOpen = !relatedOpen"
                >
                    <ChevronUp v-if="relatedOpen" class="size-5" />
                    <ChevronDown v-else class="size-5" />
                </IconButton>
            </div>
            <div
                v-if="relatedFood"
                class="flex items-center justify-between gap-2 rounded-control bg-surface-2 px-3 py-1"
            >
                <Badge tone="accent">Related</Badge>
                <span class="min-w-0 flex-1 truncate text-sm">{{ relatedFood.name }}</span>
                <Button variant="ghost" size="sm" @click="clearRelated">Clear</Button>
            </div>
            <div
                v-show="relatedOpen"
                id="related-search-panel"
                class="flex h-56 flex-col overflow-hidden rounded-control border border-line"
            >
                <SearchList
                    route="diet/meals/food/all"
                    :on-select="onPickRelated"
                    :display-component="FoodDisplay"
                />
            </div>
        </div>
        <Button
            type="submit"
            variant="primary"
            size="lg"
            block
            :loading="createFoodMutation.isPending.value"
        >
            Create
        </Button>
    </form>
</template>
