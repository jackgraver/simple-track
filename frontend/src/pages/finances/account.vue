<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import Chart from "~/shared/chart/Chart.vue";
import { chartTheme } from "~/shared/chart/theme";
import { AlertTriangle, ArrowDownLeft, ArrowUpRight, Trash2 } from "lucide-vue-next";
import {
    createTransaction,
    deleteAccount,
    setBalance,
    updateAccount,
} from "~/api/money/api";
import {
    useAccount,
    useBalances,
    useContributionTypes,
    useDeleteBalance,
    useDeleteTransaction,
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
    changeClass,
    yearAllowance,
} from "./format";
import {
    Badge,
    Button,
    Card,
    Checkbox,
    EmptyState,
    ErrorState,
    Field,
    IconButton,
    Input,
    List,
    ListRow,
    Section,
    SegmentedControl,
    Skeleton,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";

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

const showAccountSkeleton = useDelayedFlag(accountQuery.isPending);
const showBalancesSkeleton = useDelayedFlag(balancesQuery.isPending);
const showTransactionsSkeleton = useDelayedFlag(transactionsQuery.isPending);
const kindOptions = [
    { label: "Deposit", value: "deposit" as const },
    { label: "Withdrawal", value: "withdrawal" as const },
];

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
        account.value?.is_liability ? chartTheme().danger : chartTheme().success,
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
    "Failed to record balance",
);

async function saveBalance() {
    const balance = parseAmount(balanceForm.amount);
    if (balance === null) return;
    try {
        await balanceMutation.mutateAsync({ balance, date: balanceForm.date });
        toast.push("Balance recorded", "success");
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

const deleteBalanceMutation = useDeleteBalance();

async function removeBalance(balanceId: number, date: string) {
    const confirmed = await dialogManager.confirm({
        title: "Delete balance?",
        tone: "danger",
        message: `The balance recorded on ${formatDay(date)} will be removed from the history.`,
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteBalanceMutation.mutateAsync({ accountId: accountId.value, id: balanceId });
    } catch {
        // useMoneyMutation already toasted the error.
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
    // Cash moving in or out changes the balance by the same amount; the server
    // saves the transaction and the new balance together or not at all.
    await createTransaction(accountId.value, {
        kind: txForm.kind,
        amount,
        date: txForm.date,
        note: txForm.note,
        update_balance: txForm.alsoUpdateBalance && canAdjustBalance.value,
    });
}, "Failed to add transaction");

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
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

const deleteTxMutation = useDeleteTransaction();

async function removeTransaction(transactionId: number) {
    const confirmed = await dialogManager.confirm({
        title: "Delete transaction?",
        tone: "danger",
        message: "This only removes the record; the account balance is not changed.",
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteTxMutation.mutateAsync({ accountId: accountId.value, id: transactionId });
    } catch {
        // useMoneyMutation already toasted the error.
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
    "Failed to update account",
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
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

const deleteMutation = useMoneyMutation(deleteAccount, "Failed to delete account");

async function removeAccount() {
    const a = account.value;
    if (!a) return;
    const confirmed = await dialogManager.confirm({
        title: `Delete ${a.name}?`,
        tone: "danger",
        message: "All balances and transactions are permanently deleted and removed from net worth history. Archive instead to keep them.",
        confirmText: "Delete forever",
    });
    if (!confirmed) return;
    try {
        await deleteMutation.mutateAsync(a.ID);
        toast.push("Account deleted", "success");
        void router.push({ name: "finances" });
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

const navLinks = computed<ModuleTitleLink[]>(() => [
    { name: "Edit", onClick: editAccount },
]);
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar :title="account?.name ?? 'Account'" :links="navLinks" />

        <Card v-if="accountQuery.error.value" padding="none">
            <ErrorState
                :error="accountQuery.error.value"
                :title="getApiErrorMessage(accountQuery.error.value, 'Couldn\'t load this account')"
                @retry="accountQuery.refetch()"
            />
        </Card>
        <div v-else-if="!account" class="flex flex-col gap-6" aria-busy="true">
            <Skeleton v-if="showAccountSkeleton" shape="block" height="12rem" />
        </div>

        <template v-else>
            <Card v-if="account.archived_at" padding="sm" class="text-sm text-fg-muted">
                Archived {{ formatAge(account.archived_at) }}. Not counted toward net worth.
            </Card>

            <!-- Summary -->
            <Card>
              <div class="flex flex-col gap-4">
                <div class="flex flex-wrap items-start justify-between gap-4">
                    <div>
                        <p class="m-0 text-sm text-fg-muted">
                            {{ accountKindLabels[account.kind] }}<template v-if="account.institution"> · {{ account.institution }}</template>
                        </p>
                        <p class="m-0 text-3xl font-semibold text-fg tabular-nums">{{ formatMoney(account.current_balance) }}</p>
                        <p class="m-0 flex items-center gap-1.5 text-sm" :class="isStale ? 'text-warn' : 'text-fg-muted'">
                            <AlertTriangle v-if="isStale" aria-hidden="true" class="size-4" />
                            {{ account.is_liability ? "Owed" : "Balance" }}
                            {{ account.balance_date ? `as of ${formatDay(account.balance_date)}` : "not recorded yet" }}
                        </p>
                    </div>
                    <dl v-if="showsGain" class="m-0 grid grid-cols-2 gap-x-6 gap-y-2 text-sm lg:grid-cols-3">
                        <div>
                            <dt class="text-fg-muted">Deposited</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(account.total_deposits) }}</dd>
                        </div>
                        <div v-if="account.total_withdrawals > 0">
                            <dt class="text-fg-muted">Withdrawn</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(account.total_withdrawals) }}</dd>
                        </div>
                        <div>
                            <dt class="text-fg-muted">Gain</dt>
                            <dd class="m-0 font-medium tabular-nums" :class="changeClass(account.gain >= 0)">
                                {{ formatSignedMoney(account.gain) }}
                                <span v-if="gainPercent !== null" class="text-xs">({{ gainPercent.toFixed(1) }}%)</span>
                            </dd>
                        </div>
                    </dl>
                </div>

                <RouterLink
                    v-if="linkedType"
                    :to="{ name: 'finance-rules' }"
                    class="flex min-h-touch flex-wrap items-center justify-between gap-x-3 gap-y-1 rounded-control bg-surface-2 px-3 py-2 text-sm text-fg no-underline hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent lg:min-h-row"
                >
                    <span>Counts toward <strong>{{ linkedType.name }}</strong> room</span>
                    <span v-if="linkedType.room" class="text-fg-muted tabular-nums">
                        <template v-if="linkedAllowance">
                            <span :class="linkedAllowance.left < 0 ? 'text-danger' : ''">{{ formatMoney(linkedAllowance.left) }} left in {{ linkedType.room.years[0]?.year }}</span>
                            · {{ formatMoney(linkedType.room.available) }} total
                        </template>
                        <span v-else :class="linkedType.room.available < 0 ? 'text-danger' : ''">{{ formatMoney(linkedType.room.available) }} available</span>
                    </span>
                    <span v-else class="text-warn">{{ linkedType.room_issue }}</span>
                </RouterLink>

                <form class="flex flex-col gap-3 lg:flex-row lg:items-end" @submit.prevent="saveBalance">
                    <div class="lg:flex-1">
                        <Field :label="account.is_liability ? 'Amount owed' : 'New balance'" required>
                            <Input v-model="balanceForm.amount" inputmode="decimal" required />
                        </Field>
                    </div>
                    <div class="lg:w-44">
                        <Field label="Date" required>
                            <Input v-model="balanceForm.date" type="date" required />
                        </Field>
                    </div>
                    <Button type="submit" variant="primary" :loading="balanceMutation.isPending.value">Record balance</Button>
                </form>
                <p v-if="account.notes" class="m-0 text-sm whitespace-pre-line text-fg-muted">{{ account.notes }}</p>
              </div>
            </Card>

            <!-- Balance history -->
            <Section title="Balance history">
                <Card v-if="balances.length > 1">
                    <div class="h-56 w-full min-w-0">
                        <Chart type="line" :data="balanceChart.data" :options="balanceChart.options" class="h-full w-full" />
                    </div>
                </Card>
                <Card padding="none">
                    <ErrorState
                        v-if="balancesQuery.isError.value"
                        :error="balancesQuery.error.value"
                        title="Couldn't load balances"
                        @retry="balancesQuery.refetch()"
                    />
                    <SkeletonList v-else-if="showBalancesSkeleton" :rows="3" :leading="false" />
                    <EmptyState
                        v-else-if="!balancesQuery.isPending.value && balances.length === 0"
                        title="No balances recorded yet"
                        description="Record one above to start the history."
                    />
                    <List v-else-if="balances.length">
                        <ListRow
                            v-for="(balance, index) in visibleBalances"
                            :key="balance.ID"
                        >
                            <span class="text-fg-muted">{{ formatDay(balance.date) }}</span>
                            <template #trailing>
                                <span
                                    v-if="balances[index + 1] && balance.balance !== balances[index + 1].balance"
                                    class="text-xs"
                                    :class="changeClass((balance.balance - balances[index + 1].balance > 0) !== account.is_liability)"
                                >
                                    {{ formatSignedMoney(balance.balance - balances[index + 1].balance) }}
                                </span>
                                <span class="min-w-24 text-right font-medium text-fg">{{ formatMoney(balance.balance) }}</span>
                                <IconButton label="Delete balance" size="sm" variant="danger" @click="removeBalance(balance.ID, balance.date)">
                                    <Trash2 />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                </Card>
                <div v-if="balances.length > LIST_PREVIEW">
                    <Button variant="ghost" size="sm" @click="showAllBalances = !showAllBalances">
                        {{ showAllBalances ? "Show fewer" : `Show all ${balances.length}` }}
                    </Button>
                </div>
            </Section>

            <!-- Transactions -->
            <Section v-if="!account.is_liability" title="Deposits & withdrawals">
                <p class="m-0 -mt-2 text-xs text-fg-muted">
                    Money you put in or took out, used to separate contributions from growth{{ linkedType ? ` and to track ${linkedType.name} room` : "" }}.
                </p>
                <Card>
                    <form class="flex flex-col gap-3" @submit.prevent="addTransaction">
                        <div class="lg:w-64">
                            <SegmentedControl v-model="txForm.kind" :options="kindOptions" label="Transaction type" />
                        </div>
                        <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_10rem_1.5fr_auto] lg:items-end">
                            <Field label="Amount" required>
                                <Input v-model="txForm.amount" inputmode="decimal" required placeholder="0.00" />
                            </Field>
                            <Field label="Date" required>
                                <Input v-model="txForm.date" type="date" required />
                            </Field>
                            <Field label="Note">
                                <Input v-model="txForm.note" placeholder="Optional" />
                            </Field>
                            <Button type="submit" variant="primary" :loading="txMutation.isPending.value">
                                Add {{ txForm.kind }}
                            </Button>
                        </div>
                        <Checkbox
                            v-if="canAdjustBalance"
                            v-model="txForm.alsoUpdateBalance"
                            :label="`Also ${txForm.kind === 'deposit' ? 'add to' : 'subtract from'} the balance (${formatMoney(account.current_balance)})`"
                        />
                    </form>
                </Card>

                <Card padding="none">
                    <ErrorState
                        v-if="transactionsQuery.isError.value"
                        :error="transactionsQuery.error.value"
                        title="Couldn't load transactions"
                        @retry="transactionsQuery.refetch()"
                    />
                    <SkeletonList v-else-if="showTransactionsSkeleton" :rows="3" />
                    <EmptyState
                        v-else-if="!transactionsQuery.isPending.value && transactions.length === 0"
                        title="No deposits or withdrawals recorded"
                    />
                    <List v-else-if="transactions.length">
                        <ListRow
                            v-for="tx in visibleTransactions"
                            :key="tx.ID"
                        >
                            <template #leading>
                                <component
                                    :is="tx.kind === 'deposit' ? ArrowDownLeft : ArrowUpRight"
                                    aria-hidden="true"
                                    class="size-4"
                                    :class="tx.kind === 'deposit' ? 'text-success' : 'text-warn'"
                                />
                            </template>
                            {{ formatDay(tx.date) }}
                            <template v-if="tx.note" #subtitle>{{ tx.note }}</template>
                            <template #trailing>
                                <Badge :tone="tx.kind === 'deposit' ? 'success' : 'warn'">{{ tx.kind === "deposit" ? "Deposit" : "Withdrawal" }}</Badge>
                                <span class="min-w-24 text-right font-medium text-fg">{{ tx.kind === "deposit" ? "+" : "−" }}{{ formatMoney(tx.amount) }}</span>
                                <IconButton label="Delete transaction" size="sm" variant="danger" @click="removeTransaction(tx.ID)">
                                    <Trash2 />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                </Card>
                <div v-if="transactions.length > LIST_PREVIEW">
                    <Button variant="ghost" size="sm" @click="showAllTransactions = !showAllTransactions">
                        {{ showAllTransactions ? "Show fewer" : `Show all ${transactions.length}` }}
                    </Button>
                </div>
            </Section>

            <!-- Danger zone -->
            <section class="flex flex-col gap-2 border-t border-line pt-4 lg:flex-row lg:justify-end">
                <Button :loading="archiveMutation.isPending.value" @click="toggleArchive">
                    {{ account.archived_at ? "Restore account" : "Archive account" }}
                </Button>
                <Button variant="danger" :loading="deleteMutation.isPending.value" @click="removeAccount">
                    Delete account
                </Button>
            </section>
        </template>
    </div>
</template>
