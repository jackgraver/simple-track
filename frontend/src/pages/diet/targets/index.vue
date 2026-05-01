<script setup lang="ts">
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { updatePlanMacros } from "~/api/diet/api";
import { toast } from "~/composables/toast/useToast";
import { homeKeys } from "~/pages/home/queries/keys";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";

const router = useRouter();
const queryClient = useQueryClient();
const { data: dayData, isPending, error: loadError } = useDietLogsToday(0);

const calories = ref(0);
const protein = ref(0);
const fiber = ref(0);
const carbs = ref(0);

watch(
    () => dayData.value?.day.plan,
    (plan) => {
        if (!plan) return;
        calories.value = plan.calories;
        protein.value = plan.protein;
        fiber.value = plan.fiber;
        carbs.value = plan.carbs;
    },
    { immediate: true },
);

const planId = computed(() => dayData.value?.day.plan.ID);
const macrosValid = computed(() => {
    const nums = [calories.value, protein.value, fiber.value, carbs.value];
    return nums.every(
        (n) => typeof n === "number" && !Number.isNaN(n) && n >= 0,
    );
});

const saveMutation = useMutation({
    mutationFn: async () => {
        const id = planId.value;
        if (id == null) throw new Error("No diet plan loaded");
        return updatePlanMacros(id, {
            calories: calories.value,
            protein: protein.value,
            fiber: fiber.value,
            carbs: carbs.value,
        });
    },
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: homeKeys.diet.all });
        toast.push("Macro targets saved", "success");
        router.push({ name: "diet" });
    },
    onError: () => {
        toast.push("Could not save macro targets", "error");
    },
});

const saving = computed(() => saveMutation.isPending.value);
const submit = () => saveMutation.mutate();

const goBack = () => {
    router.push({ name: "diet" });
};
</script>

<template>
    <div class="mx-auto flex max-w-lg flex-col gap-6 pb-8 pt-2">
        <div class="flex flex-wrap items-center gap-3">
            <button
                type="button"
                class="text-sm text-zinc-400 underline hover:text-zinc-200"
                @click="goBack"
            >
                ← Diet
            </button>
            <h1 class="m-0 text-xl font-semibold tracking-tight text-zinc-100">
                Macro targets
            </h1>
        </div>
        <section>
            <input type="text" />
            <input type="text" />
            <h1>Suggested targets</h1>
            <p>Calories - 2000</p>
            <p>Protein - 100g</p>
            <p>Carbs - 100g</p>
            <p>Fiber - 10g</p>
        </section>
        <div v-if="isPending" class="text-zinc-500">Loading…</div>
        <div v-else-if="loadError" class="text-red-400">
            {{ loadError.message }}
        </div>
        <form v-else class="flex flex-col gap-4" @submit.prevent="submit">
            <p v-if="dayData?.day.plan?.name" class="m-0 text-sm text-zinc-400">
                Plan: {{ dayData.day.plan.name }}
            </p>
            <label class="flex flex-col gap-1">
                <span class="text-sm font-medium text-zinc-300">Calories</span>
                <input
                    v-model.number="calories"
                    type="number"
                    min="0"
                    step="1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    required
                />
            </label>
            <label class="flex flex-col gap-1">
                <span class="text-sm font-medium text-zinc-300"
                    >Protein (g)</span
                >
                <input
                    v-model.number="protein"
                    type="number"
                    min="0"
                    step="0.1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    required
                />
            </label>
            <label class="flex flex-col gap-1">
                <span class="text-sm font-medium text-zinc-300">Carbs (g)</span>
                <input
                    v-model.number="carbs"
                    type="number"
                    min="0"
                    step="0.1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    required
                />
            </label>
            <label class="flex flex-col gap-1">
                <span class="text-sm font-medium text-zinc-300">Fiber (g)</span>
                <input
                    v-model.number="fiber"
                    type="number"
                    min="0"
                    step="0.1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    required
                />
            </label>
            <button
                type="submit"
                class="rounded-md border border-amber-700/50 bg-amber-950/40 px-3 py-2 text-sm font-medium text-amber-100 hover:bg-amber-950/70 disabled:opacity-50"
                :disabled="saving || !planId || !macrosValid"
            >
                {{ saving ? "Saving…" : "Save targets" }}
            </button>
        </form>
    </div>
</template>
