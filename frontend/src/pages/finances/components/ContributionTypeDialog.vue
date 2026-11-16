<script setup lang="ts">
import { computed, reactive } from "vue";
import { createContributionType, updateContributionType } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { ContributionType, ContributionTypeSettings } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { parseAmount, ui } from "../format";

const props = defineProps<{
    contributionType?: ContributionType;
    onResolve?: (saved: boolean) => void;
    onCancel?: () => void;
}>();

type CarryMode = "unlimited" | "none" | "capped";

const existing = props.contributionType;
const str = (value: number | null | undefined) => (value === null || value === undefined ? "" : String(value));

const form = reactive({
    name: existing?.name ?? "",
    startYear: str(existing?.start_year),
    eligibilityAge: str(existing?.eligibility_age),
    carryMode: (existing?.carry_forward_cap === null || existing?.carry_forward_cap === undefined
        ? "unlimited"
        : existing.carry_forward_cap === 0
          ? "none"
          : "capped") as CarryMode,
    carryCap: existing?.carry_forward_cap ? String(existing.carry_forward_cap) : "",
    lifetimeLimit: str(existing?.lifetime_limit),
    withdrawalsRestoreRoom: existing?.withdrawals_restore_room ?? false,
    openingRoomYear: str(existing?.opening_room_year),
    openingRoom: str(existing?.opening_room),
});

const hasOpening = computed(() => form.openingRoomYear !== "" || form.openingRoom !== "");

function optionalInt(value: string) {
    const parsed = parseAmount(value);
    return parsed === null ? null : Math.trunc(parsed);
}

function settings(): ContributionTypeSettings {
    return {
        name: form.name.trim(),
        start_year: optionalInt(form.startYear),
        eligibility_age: optionalInt(form.eligibilityAge),
        carry_forward_cap:
            form.carryMode === "unlimited" ? null : form.carryMode === "none" ? 0 : parseAmount(form.carryCap),
        lifetime_limit: parseAmount(form.lifetimeLimit),
        withdrawals_restore_room: form.withdrawalsRestoreRoom,
        opening_room_year: optionalInt(form.openingRoomYear),
        opening_room: parseAmount(form.openingRoom),
    };
}

const mutation = useMoneyMutation(async (payload: ContributionTypeSettings) => {
    if (existing) await updateContributionType(existing.ID, payload);
    else await createContributionType(payload);
});

async function submit() {
    if (!form.name.trim() || mutation.isPending.value) return;
    try {
        await mutation.mutateAsync(settings());
        toast.push(existing ? "Rules saved" : "Account type added", "success");
        props.onResolve?.(true);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save rules"), "error");
    }
}
</script>

<template>
    <form class="flex w-full min-w-72 flex-col gap-5 sm:min-w-[28rem]" @submit.prevent="submit">
        <label :class="ui.label">
            Name
            <input v-model="form.name" required placeholder="e.g. TFSA" :class="ui.input" />
        </label>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-textPrimary">When room starts</legend>
            <div class="grid gap-3 sm:grid-cols-2">
                <label :class="ui.label">
                    Start year
                    <input v-model="form.startYear" inputmode="numeric" placeholder="e.g. 2023" :class="ui.input" />
                </label>
                <label :class="ui.label">
                    Eligible from age
                    <input v-model="form.eligibilityAge" inputmode="numeric" placeholder="e.g. 18" :class="ui.input" />
                </label>
            </div>
            <p class="m-0 text-xs text-textSecondary">
                Room starts at the later of the start year and the year you reach the eligibility age (needs your birth year).
                With neither, it starts at the first annual limit.
            </p>
        </fieldset>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-textPrimary">How room behaves</legend>
            <label :class="ui.label">
                Unused room
                <select v-model="form.carryMode" :class="ui.input">
                    <option value="unlimited">Carries forward forever</option>
                    <option value="capped">Carries forward up to a cap</option>
                    <option value="none">Is lost at year end</option>
                </select>
            </label>
            <label v-if="form.carryMode === 'capped'" :class="ui.label">
                Carry-forward cap
                <input v-model="form.carryCap" inputmode="decimal" required placeholder="e.g. 8000" :class="ui.input" />
            </label>
            <label :class="ui.label">
                Lifetime contribution limit
                <input v-model="form.lifetimeLimit" inputmode="decimal" placeholder="None" :class="ui.input" />
            </label>
            <label class="flex items-start gap-2 text-sm text-textPrimary">
                <input v-model="form.withdrawalsRestoreRoom" type="checkbox" class="mt-1" />
                <span>
                    Withdrawals are added back as room next January
                    <span class="block text-xs text-textSecondary">On for TFSA, off for RRSP/FHSA.</span>
                </span>
            </label>
        </fieldset>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-textPrimary">Opening room (optional)</legend>
            <p class="m-0 text-xs text-textSecondary">
                Know your room from CRA My Account or a notice of assessment? Enter it here and earlier history is ignored —
                no need to backfill years of deposits.
            </p>
            <div class="grid gap-3 sm:grid-cols-2">
                <label :class="ui.label">
                    Room available on Jan 1 of
                    <input v-model="form.openingRoomYear" inputmode="numeric" :required="hasOpening" placeholder="Year" :class="ui.input" />
                </label>
                <label :class="ui.label">
                    Amount
                    <input v-model="form.openingRoom" inputmode="decimal" :required="hasOpening" placeholder="0.00" :class="ui.input" />
                </label>
            </div>
        </fieldset>

        <div class="flex justify-end gap-2">
            <button type="button" :class="ui.btnSecondary" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" :class="ui.btnPrimary" :disabled="mutation.isPending.value">
                {{ existing ? "Save rules" : "Add type" }}
            </button>
        </div>
    </form>
</template>
