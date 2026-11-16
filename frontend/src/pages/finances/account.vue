<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import Chart from "primevue/chart";
import { ArrowDownLeft, ArrowUpRight, Trash2 } from "lucide-vue-next";
import {
    createTransaction,
    deleteAccount,
    deleteBalance,
    deleteTransaction,
    setBalance,
    updateAccount,
} from "~/api/money/api";
import {
    useAccount,
    useBalances,
    useContributionTypes,
    useMoneyMutation,
    useTransactions,
} from "~/api/money/queries";
import type { Account, TransactionKind } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import ModuleTitleBar, { type ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";
import AccountFormDialog from "./components/AccountFormDialog.vue";
import { buildMoneyLineChart } from "./chart";
import {
    STALE_BALANCE_DAYS,
    accountKindLabels,
    daysSince,
    formatAge,
    formatDay,
    formatMoney,
    formatSignedMoney,
    parseAmount,
    todayKey,
    ui,
    yearAllowance,
} from "./format";

const route = useRoute();
const router = useRouter();
const accountId = computed(() => Number(route.params.id));

const accountQuery = useAccount(accountId);
const balancesQuery = useBalances(accountId);
const transactionsQuery = useTransactions(accountId);
const typesQuery = useContributionTypes();

const account = computed(() => accountQuery.data.value);
const balances = computed(() => balancesQuery.data.value ?? []);
const transactions = computed(() => transactionsQuery.data.value ?? []);
const contributionTypes = computed(() => typesQuery.data.value?.contribution_types ?? []);
const linkedType = computed(() =>
    contributionTypes.value.find((t) => t.ID === account.value?.contribution_type_id),
);
const linkedAllowance = computed(() => {
    const year = linkedType.value?.room?.years[0];
    return year ? yearAllowance(year) : null;
});

const LIST_PREVIEW = 8;
const showAllBalances = ref(false);
const showAllTransactions = ref(false);
const visibleBalances = computed(() =>
    showAllBalances.value ? balances.value : balances.value.slice(0, LIST_PREVIEW),
);
const visibleTransactions = computed(() =>
    showAllTransactions.value ? transactions.value : transactions.value.slice(0, LIST_PREVIEW),
);

const isStale = computed(
    () => !account.value?.balance_date || daysSince(account.value.balance_date) > STALE_BALANCE_DAYS,
);
const showsGain = computed(
    () => !!account.value && !account.value.is_liability && account.value.transaction_count > 0,
);
const gainPercent = computed(() => {
    const a = account.value;
    if (!a || a.net_contributions <= 0) return null;
    return (a.gain / a.net_contributions) * 100;
});

const balanceChart = computed(() =>
    buildMoneyLineChart(
        [...balances.value].reverse().map((b) => ({ date: b.date, value: b.balance })),
        account.value?.is_liability ? "Owed" : "Balance",
        account.value?.is_liability ? "rgb(248, 113, 113)" : "rgb(52, 211, 153)",
    ),
);

// ---- Balance form ----

const balanceForm = reactive({ amount: "", date: todayKey() });
watch(
    account,
    (a) => {
        if (a && balanceForm.amount === "") balanceForm.amount = String(a.current_balance);
    },
    { immediate: true },
);

const balanceMutation = useMoneyMutation((args: { balance: number; date: string }) =>
    setBalance(accountId.value, args),
);

async function saveBalance() {
    const balance = parseAmount(balanceForm.amount);
    if (balance === null) return;
    try {
        await balanceMutation.mutateAsync({ balance, date: balanceForm.date });
        toast.push("Balance recorded", "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to record balance"), "error");
    }
}

const deleteBalanceMutation = useMoneyMutation((balanceId: number) =>
    deleteBalance(accountId.value, balanceId),
);

async function removeBalance(balanceId: number, date: string) {
    const confirmed = await dialogManager.confirm({
        title: "Delete balance?",
        message: `The balance recorded on ${formatDay(date)} will be removed from the history.`,
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteBalanceMutation.mutateAsync(balanceId);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to delete balance"), "error");
    }
}

// ---- Transactions ----

const txForm = reactive({
    kind: "deposit" as TransactionKind,
    amount: "",
    date: todayKey(),
    note: "",
    alsoUpdateBalance: true,
});

const txMutation = useMoneyMutation(async () => {
    const amount = parseAmount(txForm.amount);
    if (amount === null) return;
    await createTransaction(accountId.value, {
        kind: txForm.kind,
        amount,
        date: txForm.date,
        note: txForm.note,
    });
    // Cash moving in or out changes the balance by the same amount.
    if (txForm.alsoUpdateBalance && canAdjustBalance.value && account.value) {
        const delta = txForm.kind === "deposit" ? amount : -amount;
        await setBalance(accountId.value, {
            balance: Math.round((account.value.current_balance + delta) * 100) / 100,
            date: txForm.date,
        });
    }
});

/** Only offer to adjust the balance when the transaction is on/after the latest balance. */
const canAdjustBalance = computed(
    () => !account.value?.balance_date || txForm.date >= account.value.balance_date.slice(0, 10),
);

async function addTransaction() {
    if (parseAmount(txForm.amount) === null) return;
    try {
        await txMutation.mutateAsync(undefined);
        toast.push(txForm.kind === "deposit" ? "Deposit added" : "Withdrawal added", "success");
        txForm.amount = "";
        txForm.note = "";
        balanceForm.amount = "";
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to add transaction"), "error");
    }
}

const deleteTxMutation = useMoneyMutation((transactionId: number) =>
    deleteTransaction(accountId.value, transactionId),
);

async function removeTransaction(transactionId: number) {
    const confirmed = await dialogManager.confirm({
        title: "Delete transaction?",
        message: "This only removes the record; the account balance is not changed.",
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteTxMutation.mutateAsync(transactionId);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to delete transaction"), "error");
    }
}

// ---- Account actions ----

function editAccount() {
    if (!account.value) return;
    void dialogManager.custom({
        title: "Edit account",
        component: AccountFormDialog,
        componentProps: { account: account.value, contributionTypes: contributionTypes.value },
    });
}

const archiveMutation = useMoneyMutation((a: Account) =>
    updateAccount(a.ID, {
        name: a.name,
        kind: a.kind,
        institution: a.institution,
        notes: a.notes,
        contribution_type_id: a.contribution_type_id,
        archived: !a.archived_at,
    }),
);

async function toggleArchive() {
    const a = account.value;
    if (!a) return;
    if (!a.archived_at) {
        const confirmed = await dialogManager.confirm({
            title: `Archive ${a.name}?`,
            message: "It will be hidden and stop counting toward net worth from today. History is kept.",
            confirmText: "Archive",
        });
        if (!confirmed) return;
    }
    try {
        await archiveMutation.mutateAsync(a);
        toast.push(a.archived_at ? "Account restored" : "Account archived", "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to update account"), "error");
    }
}

const deleteMutation = useMoneyMutation(deleteAccount);

async function removeAccount() {
    const a = account.value;
    if (!a) return;
    const confirmed = await dialogManager.confirm({
        title: `Delete ${a.name}?`,
        message: "All balances and transactions are permanently deleted and removed from net worth history. Archive instead to keep them.",
        confirmText: "Delete forever",
    });
    if (!confirmed) return;
    try {
        await deleteMutation.mutateAsync(a.ID);
        toast.push("Account deleted", "success");
        void router.push({ name: "finances" });
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to delete account"), "error");
    }
}

const navLinks = computed<ModuleTitleLink[]>(() => [
    { name: "All accounts", to: "finances" },
    { name: "Edit", onClick: editAccount },
]);
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar :title="account?.name ?? 'Account'" :links="navLinks" />

        <p v-if="accountQuery.error.value" class="m-0 text-sm text-(--color-cf-red)">
            {{ getApiErrorMessage(accountQuery.error.value, "Failed to load account") }}
        </p>
        <p v-else-if="!account" class="m-0 text-sm text-textSecondary">Loading account…</p>

        <template v-else>
            <p v-if="account.archived_at" class="m-0 rounded-lg bg-secondBg px-4 py-2 text-sm text-textSecondary">
                Archived {{ formatAge(account.archived_at) }} — not counted toward net worth.
            </p>

            <!-- Summary -->
            <section :class="[ui.card, 'flex flex-col gap-4']">
                <div class="flex flex-wrap items-start justify-between gap-4">
                    <div>
                        <p class="m-0 text-sm text-textSecondary">
                            {{ accountKindLabels[account.kind] }}<template v-if="account.institution"> · {{ account.institution }}</template>
                        </p>
                        <p class="m-0 text-3xl font-semibold tabular-nums text-textPrimary">{{ formatMoney(account.current_balance) }}</p>
                        <p class="m-0 text-sm" :class="isStale ? 'text-amber-400' : 'text-textSecondary'">
                            {{ account.is_liability ? "Owed" : "Balance" }}
                            {{ account.balance_date ? `as of ${formatDay(account.balance_date)}` : "not recorded yet" }}
                        </p>
                    </div>
                    <dl v-if="showsGain" class="m-0 grid grid-cols-2 gap-x-6 gap-y-2 text-sm sm:grid-cols-3">
                        <div>
                            <dt class="text-textSecondary">Deposited</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(account.total_deposits) }}</dd>
                        </div>
                        <div v-if="account.total_withdrawals > 0">
                            <dt class="text-textSecondary">Withdrawn</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(account.total_withdrawals) }}</dd>
                        </div>
                        <div>
                            <dt class="text-textSecondary">Gain</dt>
                            <dd class="m-0 font-medium tabular-nums" :class="account.gain >= 0 ? ui.positive : ui.negative">
                                {{ formatSignedMoney(account.gain) }}
                                <span v-if="gainPercent !== null" class="text-xs">({{ gainPercent.toFixed(1) }}%)</span>
                            </dd>
                        </div>
                    </dl>
                </div>

                <router-link
                    v-if="linkedType"
                    :to="{ name: 'finance-rules' }"
                    class="flex flex-wrap items-center justify-between gap-2 rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg"
                >
                    <span>Counts toward <strong>{{ linkedType.name }}</strong> room</span>
                    <span v-if="linkedType.room" class="tabular-nums text-textSecondary">
                        <template v-if="linkedAllowance">
                            <span :class="linkedAllowance.left < 0 ? ui.negative : ''">{{ formatMoney(linkedAllowance.left) }} left in {{ linkedType.room.years[0]?.year }}</span>
                            · {{ formatMoney(linkedType.room.available) }} total
                        </template>
                        <span v-else :class="linkedType.room.available < 0 ? ui.negative : ''">{{ formatMoney(linkedType.room.available) }} available</span>
                    </span>
                    <span v-else class="text-amber-300">{{ linkedType.room_issue }}</span>
                </router-link>

                <form class="flex flex-wrap items-end gap-2" @submit.prevent="saveBalance">
                    <label :class="[ui.label, 'min-w-32 flex-1']">
                        {{ account.is_liability ? "Amount owed" : "New balance" }}
                        <input v-model="balanceForm.amount" inputmode="decimal" required :class="[ui.input, 'tabular-nums']" @focus="($event.target as HTMLInputElement).select()" />
                    </label>
                    <label :class="[ui.label, 'w-40']">
                        Date
                        <input v-model="balanceForm.date" type="date" required :class="ui.input" />
                    </label>
                    <button type="submit" :class="ui.btnPrimary" :disabled="balanceMutation.isPending.value">Record balance</button>
                </form>
                <p v-if="account.notes" class="m-0 whitespace-pre-line text-sm text-textSecondary">{{ account.notes }}</p>
            </section>

            <!-- Balance history -->
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold text-textPrimary">Balance history</h2>
                <div v-if="balances.length > 1" :class="[ui.card, 'h-56']">
                    <Chart type="line" :data="balanceChart.data" :options="balanceChart.options" class="h-full w-full" />
                </div>
                <p v-if="balances.length === 0" class="m-0 text-sm text-textSecondary">No balances recorded yet.</p>
                <ul v-else class="m-0 flex list-none flex-col divide-y divide-(--color-border) rounded-lg border border-(--color-border) bg-firstBg p-0">
                    <li v-for="(balance, index) in visibleBalances" :key="balance.ID" class="flex items-center gap-3 px-4 py-2 text-sm">
                        <span class="flex-1 whitespace-nowrap text-textSecondary">{{ formatDay(balance.date) }}</span>
                        <span
                            v-if="balances[index + 1]"
                            class="text-xs tabular-nums"
                            :class="(balance.balance - balances[index + 1].balance > 0) !== account.is_liability ? ui.positive : ui.negative"
                        >
                            {{ balance.balance === balances[index + 1].balance ? "" : formatSignedMoney(balance.balance - balances[index + 1].balance) }}
                        </span>
                        <span class="text-right font-medium tabular-nums sm:w-28">{{ formatMoney(balance.balance) }}</span>
                        <button type="button" class="rounded p-1 text-textSecondary hover:bg-secondBg hover:text-(--color-cf-red)" aria-label="Delete balance" @click="removeBalance(balance.ID, balance.date)">
                            <Trash2 :size="14" />
                        </button>
                    </li>
                </ul>
                <button
                    v-if="balances.length > LIST_PREVIEW"
                    type="button"
                    class="self-start p-0 text-sm text-textSecondary hover:text-textPrimary"
                    @click="showAllBalances = !showAllBalances"
                >
                    {{ showAllBalances ? "Show fewer" : `Show all ${balances.length}` }}
                </button>
            </section>

            <!-- Transactions -->
            <section v-if="!account.is_liability" class="flex flex-col gap-3">
                <div>
                    <h2 class="m-0 text-base font-semibold text-textPrimary">Deposits & withdrawals</h2>
                    <p class="m-0 text-xs text-textSecondary">
                        Money you put in or took out — used to separate contributions from growth{{ linkedType ? ` and to track ${linkedType.name} room` : "" }}.
                    </p>
                </div>
                <form :class="[ui.card, 'flex flex-col gap-3']" @submit.prevent="addTransaction">
                    <div class="flex gap-1 self-start rounded-md bg-secondBg p-1 text-sm">
                        <button
                            v-for="kind in (['deposit', 'withdrawal'] as const)"
                            :key="kind"
                            type="button"
                            class="rounded px-3 py-1 capitalize transition-colors"
                            :class="txForm.kind === kind ? 'bg-thirdBg text-textPrimary' : 'text-textSecondary hover:text-textPrimary'"
                            @click="txForm.kind = kind"
                        >
                            {{ kind }}
                        </button>
                    </div>
                    <div class="grid gap-2 sm:grid-cols-[1fr_10rem_1.5fr_auto] sm:items-end">
                        <label :class="ui.label">
                            Amount
                            <input v-model="txForm.amount" inputmode="decimal" required placeholder="0.00" :class="ui.input" />
                        </label>
                        <label :class="ui.label">
                            Date
                            <input v-model="txForm.date" type="date" required :class="ui.input" />
                        </label>
                        <label :class="ui.label">
                            Note
                            <input v-model="txForm.note" placeholder="Optional" :class="ui.input" />
                        </label>
                        <button type="submit" :class="ui.btnPrimary" :disabled="txMutation.isPending.value">
                            Add {{ txForm.kind }}
                        </button>
                    </div>
                    <label v-if="canAdjustBalance" class="flex items-center gap-2 text-xs text-textSecondary">
                        <input v-model="txForm.alsoUpdateBalance" type="checkbox" />
                        Also {{ txForm.kind === "deposit" ? "add to" : "subtract from" }} the balance ({{ formatMoney(account.current_balance) }})
                    </label>
                </form>

                <p v-if="transactions.length === 0" class="m-0 text-sm text-textSecondary">No deposits or withdrawals recorded.</p>
                <ul v-else class="m-0 flex list-none flex-col divide-y divide-(--color-border) rounded-lg border border-(--color-border) bg-firstBg p-0">
                    <li v-for="tx in visibleTransactions" :key="tx.ID" class="flex items-center gap-3 px-4 py-2 text-sm">
                        <component
                            :is="tx.kind === 'deposit' ? ArrowDownLeft : ArrowUpRight"
                            :size="16"
                            class="shrink-0"
                            :class="tx.kind === 'deposit' ? ui.positive : 'text-amber-400'"
                        />
                        <div class="min-w-0 flex-1">
                            <p class="m-0 text-textPrimary">{{ formatDay(tx.date) }}</p>
                            <p v-if="tx.note" class="m-0 truncate text-xs text-textSecondary">{{ tx.note }}</p>
                        </div>
                        <span class="font-medium tabular-nums">{{ tx.kind === "deposit" ? "+" : "−" }}{{ formatMoney(tx.amount) }}</span>
                        <button type="button" class="rounded p-1 text-textSecondary hover:bg-secondBg hover:text-(--color-cf-red)" aria-label="Delete transaction" @click="removeTransaction(tx.ID)">
                            <Trash2 :size="14" />
                        </button>
                    </li>
                </ul>
                <button
                    v-if="transactions.length > LIST_PREVIEW"
                    type="button"
                    class="self-start p-0 text-sm text-textSecondary hover:text-textPrimary"
                    @click="showAllTransactions = !showAllTransactions"
                >
                    {{ showAllTransactions ? "Show fewer" : `Show all ${transactions.length}` }}
                </button>
            </section>

            <!-- Danger zone -->
            <section class="flex flex-wrap justify-end gap-2 border-t border-(--color-border) pt-4">
                <button type="button" :class="ui.btnSecondary" :disabled="archiveMutation.isPending.value" @click="toggleArchive">
                    {{ account.archived_at ? "Restore account" : "Archive account" }}
                </button>
                <button type="button" :class="ui.btnDanger" :disabled="deleteMutation.isPending.value" @click="removeAccount">
                    Delete account
                </button>
            </section>
        </template>
    </div>
</template>
