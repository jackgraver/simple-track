<script setup lang="ts">
import { computed, reactive } from "vue";
import { createAccount, updateAccount } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { Account, AccountKind, ContributionType } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import {
    accountKindLabels,
    accountKindOrder,
    isLiabilityKind,
    parseAmount,
    todayKey,
    ui,
} from "../format";

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
});

async function submit() {
    if (!form.name.trim() || mutation.isPending.value) return;
    try {
        const id = await mutation.mutateAsync(undefined);
        toast.push(isEdit.value ? "Account updated" : "Account added", "success");
        props.onResolve?.(id);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save account"), "error");
    }
}
</script>

<template>
    <form class="flex w-full min-w-72 flex-col gap-4 sm:min-w-96" @submit.prevent="submit">
        <label :class="ui.label">
            Name
            <input v-model="form.name" required placeholder="e.g. Tangerine chequing" :class="ui.input" />
        </label>
        <div class="grid gap-3 sm:grid-cols-2">
            <label :class="ui.label">
                Type
                <select v-model="form.kind" :class="ui.input">
                    <option v-for="kind in accountKindOrder" :key="kind" :value="kind">
                        {{ accountKindLabels[kind] }}
                    </option>
                </select>
            </label>
            <label :class="ui.label">
                Institution
                <input v-model="form.institution" placeholder="Optional" :class="ui.input" />
            </label>
        </div>
        <label v-if="!isLiability" :class="ui.label">
            Registered account (shares contribution room)
            <select v-model="form.contributionTypeId" :class="ui.input">
                <option :value="null">None</option>
                <option v-for="type in contributionTypes" :key="type.ID" :value="type.ID">
                    {{ type.name }}
                </option>
            </select>
            <span v-if="contributionTypes.length === 0" class="font-normal">
                Set up TFSA / RRSP / FHSA under Contribution room to link accounts.
            </span>
        </label>
        <div v-if="!isEdit" class="grid gap-3 sm:grid-cols-2">
            <label :class="ui.label">
                {{ balanceLabel }}
                <input
                    v-model="form.openingBalance"
                    inputmode="decimal"
                    placeholder="0.00"
                    :class="ui.input"
                />
            </label>
            <label :class="ui.label">
                As of
                <input v-model="form.balanceDate" type="date" required :class="ui.input" />
            </label>
        </div>
        <label :class="ui.label">
            Notes
            <textarea v-model="form.notes" rows="2" placeholder="Optional" :class="ui.input" />
        </label>
        <label v-if="isEdit" class="flex items-center gap-2 text-sm text-textSecondary">
            <input v-model="form.archived" type="checkbox" />
            Archived (closed account — hidden and excluded from net worth)
        </label>
        <div class="flex justify-end gap-2">
            <button type="button" :class="ui.btnSecondary" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" :class="ui.btnPrimary" :disabled="mutation.isPending.value">
                {{ isEdit ? "Save" : "Add account" }}
            </button>
        </div>
    </form>
</template>
