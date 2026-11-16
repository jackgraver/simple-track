<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { createContributionTypeFromPreset } from "~/api/money/api";
import { useContributionPresets, useMoneyMutation } from "~/api/money/queries";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { formatMoney, parseAmount, ui } from "../format";

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
    <form class="flex w-full min-w-72 flex-col gap-4 sm:min-w-[28rem]" @submit.prevent="submit">
        <p v-if="isPending" class="m-0 text-sm text-textSecondary">Loading presets…</p>
        <div class="grid gap-2 sm:grid-cols-3">
            <button
                v-for="preset in available"
                :key="preset.key"
                type="button"
                class="rounded-md border px-3 py-2 text-left text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                :class="form.key === preset.key ? 'border-green-600 bg-green-950/40 text-textPrimary' : 'border-(--color-border) bg-secondBg hover:bg-thirdBg'"
                :disabled="preset.taken"
                @click="form.key = preset.key"
            >
                <span class="block font-semibold">{{ preset.name }}</span>
                <span class="block text-xs text-textSecondary">{{ preset.taken ? "Already added" : "Canada" }}</span>
            </button>
        </div>

        <p v-if="!isPending && available.length && available.every((p) => p.taken)" class="m-0 text-sm text-textSecondary">
            All presets are already set up. Use <strong>Custom type</strong> for anything else, or edit an existing type's rules.
        </p>

        <template v-if="selected">
            <p class="m-0 text-sm text-textSecondary">{{ selected.description }}</p>
            <ul v-if="selected.limits.length" class="m-0 flex list-none flex-wrap gap-x-4 gap-y-1 p-0 text-xs text-textSecondary">
                <li v-for="limit in selected.limits" :key="limit.year">
                    {{ limit.year }}+: <span class="text-textPrimary">{{ formatMoney(limit.annual_limit) }}</span>
                </li>
            </ul>

            <label v-if="selected.needs_start_year" :class="ui.label">
                Year you opened your first {{ selected.name }}
                <input v-model="form.startYear" inputmode="numeric" required placeholder="e.g. 2024" :class="ui.input" />
            </label>

            <div class="flex flex-col gap-2">
                <p class="m-0 text-sm font-medium text-textPrimary">
                    Opening room {{ selected.suggest_opening_room ? "(recommended)" : "(optional)" }}
                </p>
                <p class="m-0 text-xs text-textSecondary">
                    {{
                        selected.suggest_opening_room
                            ? "Your RRSP deduction limit from your latest notice of assessment."
                            : "If you know your exact room from CRA, enter it — otherwise it's calculated from your birth year and deposits."
                    }}
                </p>
                <div class="grid gap-3 sm:grid-cols-2">
                    <label :class="ui.label">
                        Room on Jan 1 of
                        <input v-model="form.openingRoomYear" inputmode="numeric" :class="ui.input" />
                    </label>
                    <label :class="ui.label">
                        Amount
                        <input v-model="form.openingRoom" inputmode="decimal" placeholder="Leave blank to calculate" :class="ui.input" />
                    </label>
                </div>
            </div>
        </template>

        <div class="flex justify-end gap-2">
            <button type="button" :class="ui.btnSecondary" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" :class="ui.btnPrimary" :disabled="!selected || selected.taken || mutation.isPending.value">
                Add {{ selected?.name ?? "" }}
            </button>
        </div>
    </form>
</template>
