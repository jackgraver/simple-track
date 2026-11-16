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
    ui,
} from "../format";

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
    <form class="flex w-full min-w-72 flex-col gap-4 sm:min-w-[28rem]" @submit.prevent="submit">
        <p class="m-0 text-sm text-textSecondary">
            Confirm or update every balance in one go. Unchanged values are still recorded so you know they're current.
        </p>
        <label :class="ui.label">
            Balances as of
            <input v-model="date" type="date" required :class="ui.input" />
        </label>
        <div class="flex max-h-[55vh] flex-col divide-y divide-(--color-border) overflow-y-auto [scrollbar-color:var(--color-third-bg)_transparent] [scrollbar-width:thin]">
            <div
                v-for="account in accounts"
                :key="account.ID"
                class="grid grid-cols-[1fr_9rem] items-center gap-3 py-2"
            >
                <div class="min-w-0">
                    <p class="m-0 truncate text-sm font-medium text-textPrimary">{{ account.name }}</p>
                    <p class="m-0 text-xs text-textSecondary">
                        {{ accountKindLabels[account.kind] }}
                        <template v-if="account.balance_date"> · updated {{ formatAge(account.balance_date) }}</template>
                        <template v-if="change(account) !== null">
                            ·
                            <span :class="(change(account) ?? 0) > 0 !== account.is_liability ? ui.positive : ui.negative">
                                {{ (change(account) ?? 0) > 0 ? "+" : "" }}{{ formatMoney(change(account) ?? 0) }}
                            </span>
                        </template>
                    </p>
                </div>
                <input
                    v-model="values[account.ID]"
                    inputmode="decimal"
                    :aria-label="`${account.name} balance`"
                    @focus="($event.target as HTMLInputElement).select()"
                    :class="[ui.input, 'text-right tabular-nums']"
                />
            </div>
        </div>
        <div class="flex justify-end gap-2">
            <button type="button" :class="ui.btnSecondary" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" :class="ui.btnPrimary" :disabled="mutation.isPending.value || entries.length === 0">
                Record {{ entries.length }} balance{{ entries.length === 1 ? "" : "s" }}
            </button>
        </div>
    </form>
</template>
