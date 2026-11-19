<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { createContributionTypeFromPreset } from "~/api/money/api";
import { useContributionPresets, useMoneyMutation } from "~/api/money/queries";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { Button, Field, Input, Skeleton } from "~/shared/ui";
import { formatMoney, parseAmount } from "../format";

const props = defineProps<{
    existingNames: string[];
    onResolve?: (saved: boolean) => void;
    onCancel?: () => void;
}>();

const { data: presets, isPending } = useContributionPresets();

const form = reactive({
    key: "",
    startYear: "",
    openingRoomYear: String(new Date().getFullYear()),
    openingRoom: "",
});

const available = computed(() =>
    (presets.value ?? []).map((preset) => ({
        ...preset,
        taken: props.existingNames.some((n) => n.toLowerCase() === preset.name.toLowerCase()),
    })),
);
const selected = computed(() => available.value.find((p) => p.key === form.key));

watch(
    available,
    (list) => {
        if (!form.key) form.key = list.find((p) => !p.taken)?.key ?? "";
    },
    { immediate: true },
);

const mutation = useMoneyMutation(createContributionTypeFromPreset);

async function submit() {
    const preset = selected.value;
    if (!preset || preset.taken || mutation.isPending.value) return;
    const openingRoom = parseAmount(form.openingRoom);
    const startYear = parseAmount(form.startYear);
    try {
        await mutation.mutateAsync({
            key: preset.key,
            start_year: startYear === null ? null : Math.trunc(startYear),
            opening_room_year: openingRoom === null ? null : Number(form.openingRoomYear),
            opening_room: openingRoom,
        });
        toast.push(`${preset.name} added`, "success");
        props.onResolve?.(true);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to add account type"), "error");
    }
}
</script>

<template>
    <form class="flex w-full flex-col gap-4" @submit.prevent="submit">
        <div v-if="isPending" class="grid grid-cols-1 gap-2 lg:grid-cols-3" aria-busy="true">
            <Skeleton v-for="n in 3" :key="n" height="3.5rem" />
        </div>
        <div v-else role="radiogroup" aria-label="Account type" class="grid grid-cols-1 gap-2 lg:grid-cols-3">
            <button
                v-for="preset in available"
                :key="preset.key"
                type="button"
                role="radio"
                :aria-checked="form.key === preset.key"
                class="m-0! min-h-touch cursor-pointer rounded-control! border! px-3! py-2! text-left text-base shadow-none! transition-colors duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-40 lg:text-sm"
                :class="
                    form.key === preset.key
                        ? 'border-accent! bg-accent-soft! text-fg'
                        : 'border-line! bg-surface-2! text-fg hover:bg-surface-3!'
                "
                :disabled="preset.taken"
                @click="form.key = preset.key"
            >
                <span class="block font-semibold">{{ preset.name }}</span>
                <span class="block text-xs text-fg-muted">{{ preset.taken ? "Already added" : "Canada" }}</span>
            </button>
        </div>

        <p v-if="!isPending && available.length && available.every((p) => p.taken)" class="m-0 text-sm text-fg-muted">
            All presets are already set up. Use <strong>Custom type</strong> for anything else, or edit an existing type's rules.
        </p>

        <template v-if="selected">
            <p class="m-0 text-sm text-fg-muted">{{ selected.description }}</p>
            <ul v-if="selected.limits.length" class="m-0 flex list-none flex-wrap gap-x-4 gap-y-1 p-0 text-xs text-fg-muted">
                <li v-for="limit in selected.limits" :key="limit.year">
                    {{ limit.year }}+: <span class="text-fg">{{ formatMoney(limit.annual_limit) }}</span>
                </li>
            </ul>

            <Field v-if="selected.needs_start_year" :label="`Year you opened your first ${selected.name}`" required>
                <Input v-model="form.startYear" inputmode="numeric" required placeholder="e.g. 2024" />
            </Field>

            <div class="flex flex-col gap-2">
                <p class="m-0 text-sm font-medium text-fg">
                    Opening room {{ selected.suggest_opening_room ? "(recommended)" : "(optional)" }}
                </p>
                <p class="m-0 text-xs text-fg-muted">
                    {{
                        selected.suggest_opening_room
                            ? "Your RRSP deduction limit from your latest notice of assessment."
                            : "If you know your exact room from CRA, enter it. Otherwise it's calculated from your birth year and deposits."
                    }}
                </p>
                <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                    <Field label="Room on Jan 1 of">
                        <Input v-model="form.openingRoomYear" inputmode="numeric" />
                    </Field>
                    <Field label="Amount">
                        <Input v-model="form.openingRoom" inputmode="decimal" placeholder="Leave blank to calculate" />
                    </Field>
                </div>
            </div>
        </template>

        <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
            <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
            <Button
                type="submit"
                variant="primary"
                :disabled="!selected || selected.taken"
                :loading="mutation.isPending.value"
            >
                Add {{ selected?.name ?? "" }}
            </Button>
        </div>
    </form>
</template>
