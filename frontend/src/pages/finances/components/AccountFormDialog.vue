<script setup lang="ts">
import { computed, reactive } from "vue";
import { createAccount, updateAccount } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { Account, AccountKind, ContributionType } from "~/api/money/types";
import { toast } from "~/composables/toast/useToast";
import {
    accountKindLabels,
    accountKindOrder,
    isLiabilityKind,
    parseAmount,
    todayKey,
} from "../format";
import { Button, Checkbox, Field, Input, Select, Textarea } from "~/shared/ui";

const props = defineProps<{
    account?: Account;
    contributionTypes: ContributionType[];
    defaultKind?: AccountKind;
    onResolve?: (accountId: number | true) => void;
    onCancel?: () => void;
}>();

const isEdit = computed(() => !!props.account);

const form = reactive({
    name: props.account?.name ?? "",
    kind: (props.account?.kind ?? props.defaultKind ?? "chequing") as AccountKind,
    institution: props.account?.institution ?? "",
    notes: props.account?.notes ?? "",
    contributionTypeId: props.account?.contribution_type_id ?? null,
    archived: !!props.account?.archived_at,
    openingBalance: "",
    balanceDate: todayKey(),
});

const kindOptions = accountKindOrder.map((kind) => ({ label: accountKindLabels[kind], value: kind }));
// Select works on strings/numbers, so "None" is 0 rather than null.
const typeOptions = computed(() => [
    { label: "None", value: 0 },
    ...props.contributionTypes.map((t) => ({ label: t.name, value: t.ID })),
]);
const typeId = computed<number | "">({
    get: () => form.contributionTypeId ?? 0,
    set: (v) => {
        form.contributionTypeId = v === "" || v === 0 ? null : v;
    },
});

const isLiability = computed(() => isLiabilityKind(form.kind));
const balanceLabel = computed(() =>
    isLiability.value ? "Amount owed" : "Current balance",
);

const mutation = useMoneyMutation(async () => {
    const settings = {
        name: form.name.trim(),
        kind: form.kind,
        institution: form.institution.trim(),
        notes: form.notes.trim(),
        contribution_type_id: isLiability.value ? null : form.contributionTypeId,
        archived: form.archived,
    };
    if (props.account) {
        await updateAccount(props.account.ID, settings);
        return props.account.ID;
    }
    const created = await createAccount({
        ...settings,
        opening_balance: parseAmount(form.openingBalance),
        balance_date: form.balanceDate,
    });
    return created.ID;
}, "Failed to save account");

async function submit() {
    if (!form.name.trim() || mutation.isPending.value) return;
    try {
        const id = await mutation.mutateAsync(undefined);
        toast.push(isEdit.value ? "Account updated" : "Account added", "success");
        props.onResolve?.(id);
    } catch {
        // useMoneyMutation already toasted the error.
    }
}
</script>

<template>
    <form class="flex w-full flex-col gap-4" @submit.prevent="submit">
        <Field label="Name" required>
            <Input v-model="form.name" required placeholder="e.g. Tangerine chequing" />
        </Field>
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
            <Field label="Type">
                <Select v-model="form.kind" :options="kindOptions" />
            </Field>
            <Field label="Institution">
                <Input v-model="form.institution" placeholder="Optional" />
            </Field>
        </div>
        <Field
            v-if="!isLiability"
            label="Registered account (shares contribution room)"
            :hint="contributionTypes.length === 0 ? 'Set up TFSA / RRSP / FHSA under Contribution room to link accounts.' : undefined"
        >
            <Select v-model="typeId" :options="typeOptions" />
        </Field>
        <div v-if="!isEdit" class="grid grid-cols-1 gap-3 lg:grid-cols-2">
            <Field :label="balanceLabel">
                <Input v-model="form.openingBalance" inputmode="decimal" placeholder="0.00" />
            </Field>
            <Field label="As of" required>
                <Input v-model="form.balanceDate" type="date" required />
            </Field>
        </div>
        <Field label="Notes">
            <Textarea v-model="form.notes" :rows="2" placeholder="Optional" />
        </Field>
        <Checkbox
            v-if="isEdit"
            v-model="form.archived"
            label="Archived (closed account: hidden and excluded from net worth)"
        />
        <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
            <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
            <Button type="submit" variant="primary" :loading="mutation.isPending.value">
                {{ isEdit ? "Save" : "Add account" }}
            </Button>
        </div>
    </form>
</template>
