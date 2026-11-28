<script setup lang="ts">
import { computed, reactive } from "vue";
import { createContributionType, setContributionRoom, updateContributionType } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { ContributionRoomSettings, ContributionType, ContributionTypeRules } from "~/api/money/types";
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
// The rules are shared: only the creator and the admin change them. Everyone sets their own room.
const canEditRules = !existing || existing.can_edit;
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

function rules(): ContributionTypeRules {
    return {
        name: form.name.trim(),
        eligibility_age: optionalInt(form.eligibilityAge),
        carry_forward_cap:
            form.carryMode === "unlimited" ? null : form.carryMode === "none" ? 0 : parseAmount(form.carryCap),
        lifetime_limit: parseAmount(form.lifetimeLimit),
        withdrawals_restore_room: form.withdrawalsRestoreRoom,
    };
}

function room(): ContributionRoomSettings {
    return {
        start_year: optionalInt(form.startYear),
        opening_room_year: optionalInt(form.openingRoomYear),
        opening_room: parseAmount(form.openingRoom),
    };
}

const mutation = useMoneyMutation(
    async (payload: { rules: ContributionTypeRules; room: ContributionRoomSettings }) => {
        if (!existing) {
            await createContributionType({ ...payload.rules, ...payload.room });
            return;
        }
        if (canEditRules) await updateContributionType(existing.ID, payload.rules);
        await setContributionRoom(existing.ID, payload.room);
    },
    "Failed to save",
);

async function submit() {
    if (!form.name.trim() || mutation.isPending.value) return;
    try {
        await mutation.mutateAsync({ rules: rules(), room: room() });
        toast.push(!existing ? "Account type added" : canEditRules ? "Rules saved" : "Room saved", "success");
        props.onResolve?.(true);
    } catch {
        // useMoneyMutation already toasted the error.
    }
}
</script>

<template>
    <form class="flex w-full flex-col gap-5" @submit.prevent="submit">
        <p v-if="!canEditRules" class="m-0 text-sm text-fg-muted">
            {{ existing?.name }} is shared. Only its creator or the admin can change its rules; the figures below are yours alone.
        </p>
        <Field v-else label="Name" required>
            <Input v-model="form.name" required placeholder="e.g. TFSA" />
        </Field>

        <fieldset v-if="canEditRules" class="m-0 flex flex-col gap-3 border-0 p-0">
            <legend class="mb-2 text-sm font-semibold text-fg">When room starts</legend>
            <Field label="Eligible from age">
                <Input v-model="form.eligibilityAge" inputmode="numeric" placeholder="e.g. 18" />
            </Field>
            <p class="m-0 text-xs text-fg-muted">
                Room starts at the later of your own start year (below) and the year you reach the eligibility age (needs your
                birth year). With neither, it starts at the first annual limit.
            </p>
        </fieldset>

        <fieldset v-if="canEditRules" class="m-0 flex flex-col gap-3 border-0 p-0">
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
            <legend class="mb-2 text-sm font-semibold text-fg">Your room (optional)</legend>
            <Field label="Start year">
                <Input v-model="form.startYear" inputmode="numeric" placeholder="e.g. 2023" />
            </Field>
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
                {{ !existing ? "Add type" : canEditRules ? "Save rules" : "Save my room" }}
            </Button>
        </div>
    </form>
</template>
