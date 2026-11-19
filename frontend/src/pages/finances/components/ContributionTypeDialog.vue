<script setup lang="ts">
import { computed, reactive } from "vue";
import { createContributionType, updateContributionType } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { ContributionType, ContributionTypeSettings } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { Button, Checkbox, Field, Input, Select } from "~/shared/ui";
import { parseAmount } from "../format";

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

const carryOptions: { label: string; value: CarryMode }[] = [
    { label: "Carries forward forever", value: "unlimited" },
    { label: "Carries forward up to a cap", value: "capped" },
    { label: "Is lost at year end", value: "none" },
];

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
    <form class="flex w-full flex-col gap-5" @submit.prevent="submit">
        <Field label="Name" required>
            <Input v-model="form.name" required placeholder="e.g. TFSA" />
        </Field>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-fg">When room starts</legend>
            <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                <Field label="Start year">
                    <Input v-model="form.startYear" inputmode="numeric" placeholder="e.g. 2023" />
                </Field>
                <Field label="Eligible from age">
                    <Input v-model="form.eligibilityAge" inputmode="numeric" placeholder="e.g. 18" />
                </Field>
            </div>
            <p class="m-0 text-xs text-fg-muted">
                Room starts at the later of the start year and the year you reach the eligibility age (needs your birth year).
                With neither, it starts at the first annual limit.
            </p>
        </fieldset>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-fg">How room behaves</legend>
            <Field label="Unused room">
                <Select v-model="form.carryMode" :options="carryOptions" />
            </Field>
            <Field v-if="form.carryMode === 'capped'" label="Carry-forward cap" required>
                <Input v-model="form.carryCap" inputmode="decimal" required placeholder="e.g. 8000" />
            </Field>
            <Field label="Lifetime contribution limit">
                <Input v-model="form.lifetimeLimit" inputmode="decimal" placeholder="None" />
            </Field>
            <div class="flex flex-col">
                <Checkbox
                    v-model="form.withdrawalsRestoreRoom"
                    label="Withdrawals are added back as room next January"
                />
                <p class="m-0 pl-8 text-xs text-fg-muted">On for TFSA, off for RRSP/FHSA.</p>
            </div>
        </fieldset>

        <fieldset class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-fg">Opening room (optional)</legend>
            <p class="m-0 text-xs text-fg-muted">
                Know your room from CRA My Account or a notice of assessment? Enter it here and earlier history is ignored,
                so there's no need to backfill years of deposits.
            </p>
            <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                <Field label="Room available on Jan 1 of" :required="hasOpening">
                    <Input v-model="form.openingRoomYear" inputmode="numeric" :required="hasOpening" placeholder="Year" />
                </Field>
                <Field label="Amount" :required="hasOpening">
                    <Input v-model="form.openingRoom" inputmode="decimal" :required="hasOpening" placeholder="0.00" />
                </Field>
            </div>
        </fieldset>

        <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
            <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
            <Button type="submit" variant="primary" :loading="mutation.isPending.value">
                {{ existing ? "Save rules" : "Add type" }}
            </Button>
        </div>
    </form>
</template>
