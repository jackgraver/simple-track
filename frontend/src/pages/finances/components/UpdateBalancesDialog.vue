<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { setBalances } from "~/api/money/api";
import { useMoneyMutation } from "~/api/money/queries";
import type { Account } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import {
    accountKindLabels,
    formatAge,
    formatMoney,
    parseAmount,
    todayKey,
    changeClass,
} from "../format";
import { Button, Field, Input, List, ListRow } from "~/shared/ui";

const props = defineProps<{
    accounts: Account[];
    onResolve?: (saved: boolean) => void;
    onCancel?: () => void;
}>();

const date = ref(todayKey());
const values = reactive<Record<number, string>>(
    Object.fromEntries(props.accounts.map((a) => [a.ID, String(a.current_balance)])),
);

const entries = computed(() =>
    props.accounts.flatMap((account) => {
        const balance = parseAmount(values[account.ID]);
        return balance === null ? [] : [{ account_id: account.ID, balance }];
    }),
);

function change(account: Account) {
    const next = parseAmount(values[account.ID]);
    if (next === null) return null;
    const diff = Math.round((next - account.current_balance) * 100) / 100;
    return diff === 0 ? null : diff;
}

const mutation = useMoneyMutation(setBalances);

async function submit() {
    if (entries.value.length === 0 || mutation.isPending.value) return;
    try {
        await mutation.mutateAsync({ date: date.value, balances: entries.value });
        toast.push(`Recorded ${entries.value.length} balance(s)`, "success");
        props.onResolve?.(true);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save balances"), "error");
    }
}
</script>

<template>
    <form class="flex w-full flex-col gap-4" @submit.prevent="submit">
        <p class="m-0 text-sm text-fg-muted">
            Confirm or update every balance in one go. Unchanged values are still recorded so you know they're current.
        </p>
        <Field label="Balances as of" required>
            <Input v-model="date" type="date" required />
        </Field>
        <List class="max-h-[55vh] overflow-y-auto">
            <ListRow v-for="account in accounts" :key="account.ID">
                {{ account.name }}
                <template #subtitle>
                    {{ accountKindLabels[account.kind] }}
                    <template v-if="account.balance_date"> · updated {{ formatAge(account.balance_date) }}</template>
                    <template v-if="change(account) !== null">
                        ·
                        <span :class="changeClass((change(account) ?? 0) > 0 !== account.is_liability)">
                            {{ (change(account) ?? 0) > 0 ? "+" : "−" }}{{ formatMoney(Math.abs(change(account) ?? 0)) }}
                        </span>
                    </template>
                </template>
                <template #trailing>
                    <div class="w-36">
                        <Input
                            v-model="values[account.ID]"
                            inputmode="decimal"
                            :aria-label="`${account.name} balance`"
                            class="text-right"
                        />
                    </div>
                </template>
            </ListRow>
        </List>
        <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
            <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
            <Button
                type="submit"
                variant="primary"
                :disabled="entries.length === 0"
                :loading="mutation.isPending.value"
            >
                Record {{ entries.length }} balance{{ entries.length === 1 ? "" : "s" }}
            </Button>
        </div>
    </form>
</template>
