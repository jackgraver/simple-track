<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
    createInvestmentDeposit,
    listInvestmentAccounts,
    listInvestmentDeposits,
    type InvestmentAccount,
    type InvestmentDeposit,
} from "~/api/money/api";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

const accounts = ref<InvestmentAccount[]>([]);
const deposits = ref<Record<number, InvestmentDeposit[]>>({});
const expandedIds = ref(new Set<number>());
const depositAmounts = reactive<Record<number, string>>({});
const depositDates = reactive<Record<number, string>>({});
const isLoading = ref(true);
const isSaving = ref(false);
const errorMessage = ref("");

const totalBalance = computed(() =>
    accounts.value.reduce((total, account) => total + account.current_balance, 0),
);
const totalDeposits = computed(() =>
    accounts.value.reduce((total, account) => total + account.total_deposits, 0),
);
const totalProfit = computed(() =>
    accounts.value.reduce((total, account) => total + account.profit, 0),
);

function today() {
    return new Date().toISOString().slice(0, 10);
}

function formatMoney(value: number) {
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
    }).format(value);
}

function isExpanded(accountId: number) {
    return expandedIds.value.has(accountId);
}

async function load() {
    isLoading.value = true;
    errorMessage.value = "";
    try {
        accounts.value = await listInvestmentAccounts();
    } catch (error) {
        errorMessage.value = getApiErrorMessage(error, "Failed to load investments");
    } finally {
        isLoading.value = false;
    }
}

async function toggleDetails(accountId: number) {
    if (isExpanded(accountId)) {
        const next = new Set(expandedIds.value);
        next.delete(accountId);
        expandedIds.value = next;
        return;
    }
    try {
        deposits.value = {
            ...deposits.value,
            [accountId]: await listInvestmentDeposits(accountId),
        };
        depositDates[accountId] ??= today();
        expandedIds.value = new Set([...expandedIds.value, accountId]);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to load account details"), "error");
    }
}

async function addDeposit(accountId: number) {
    const amount = depositAmounts[accountId];
    const date = depositDates[accountId];
    if (!amount || !date || isSaving.value) return;
    isSaving.value = true;
    try {
        await createInvestmentDeposit(accountId, {
            amount: Number(amount),
            date,
        });
        depositAmounts[accountId] = "";
        deposits.value = {
            ...deposits.value,
            [accountId]: await listInvestmentDeposits(accountId),
        };
        await load();
        toast.push("Deposit added", "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to add deposit"), "error");
    } finally {
        isSaving.value = false;
    }
}

onMounted(() => {
    void load();
});
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-8">
        <header class="flex flex-wrap items-start justify-between gap-4">
            <div>
                <h1 class="m-0 text-xl font-semibold text-textPrimary">Finances</h1>
                <p class="mb-0 mt-1 text-sm text-textSecondary">Your investment account summary</p>
            </div>
            <div class="flex gap-2">
                <button type="button" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg disabled:opacity-50" :disabled="isLoading" @click="load">Refresh</button>
                <router-link :to="{ name: 'finance-config' }" class="rounded-md bg-primary px-3 py-2 text-sm text-white">Configure</router-link>
            </div>
        </header>
        <p v-if="errorMessage" class="m-0 text-sm text-(--color-cf-red)">{{ errorMessage }}</p>
        <p v-else-if="isLoading" class="m-0 text-sm text-textSecondary">Loading investments…</p>
        <template v-else>
            <section class="grid gap-3 sm:grid-cols-3">
                <div class="rounded-lg bg-secondBg p-4"><p class="m-0 text-sm text-textSecondary">Total balance</p><p class="mb-0 mt-1 text-xl font-semibold">{{ formatMoney(totalBalance) }}</p></div>
                <div class="rounded-lg bg-secondBg p-4"><p class="m-0 text-sm text-textSecondary">Total deposits</p><p class="mb-0 mt-1 text-xl font-semibold">{{ formatMoney(totalDeposits) }}</p></div>
                <div class="rounded-lg bg-secondBg p-4"><p class="m-0 text-sm text-textSecondary">Total profit</p><p class="mb-0 mt-1 text-xl font-semibold" :class="totalProfit >= 0 ? 'text-emerald-400' : 'text-(--color-cf-red)'">{{ formatMoney(totalProfit) }}</p></div>
            </section>
            <section class="flex flex-col gap-3">
                <div class="flex items-center justify-between">
                    <h2 class="m-0 text-base font-semibold">Accounts</h2>
                    <span class="text-sm text-textSecondary">{{ accounts.length }} total</span>
                </div>
                <div v-if="accounts.length === 0" class="rounded-lg bg-secondBg p-4 text-sm text-textSecondary">
                    No accounts yet. <router-link :to="{ name: 'finance-config' }" class="text-textPrimary underline">Configure your first account</router-link>.
                </div>
                <article v-for="account in accounts" :key="account.ID" class="rounded-lg bg-secondBg p-4">
                    <div class="flex flex-wrap items-start justify-between gap-3">
                        <div>
                            <h3 class="m-0 font-medium">{{ account.name }}</h3>
                            <p class="mb-0 mt-1 text-sm text-textSecondary">{{ account.investment_account_type?.name ?? "Uncategorized" }}</p>
                        </div>
                        <button type="button" class="text-sm hover:underline" @click="toggleDetails(account.ID)">{{ isExpanded(account.ID) ? "Hide details" : "View details" }}</button>
                    </div>
                    <dl class="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
                        <div><dt class="text-textSecondary">Balance</dt><dd class="m-0 font-medium">{{ formatMoney(account.current_balance) }}</dd></div>
                        <div><dt class="text-textSecondary">Deposits</dt><dd class="m-0 font-medium">{{ formatMoney(account.total_deposits) }}</dd></div>
                        <div><dt class="text-textSecondary">Profit</dt><dd class="m-0 font-medium" :class="account.profit >= 0 ? 'text-emerald-400' : 'text-(--color-cf-red)'">{{ formatMoney(account.profit) }}</dd></div>
                        <div v-if="account.contribution_room"><dt class="text-textSecondary">Contribution room</dt><dd class="m-0 font-medium">{{ formatMoney(account.contribution_room.remaining) }}</dd></div>
                    </dl>
                    <div v-if="isExpanded(account.ID)" class="mt-5 border-t border-white/10 pt-4">
                        <p v-if="account.contribution_room" class="m-0 text-sm text-textSecondary">Since {{ account.contribution_room.eligible_from_year }}: {{ formatMoney(account.contribution_room.earned_room) }} earned, {{ formatMoney(account.contribution_room.contributed) }} contributed, {{ formatMoney(account.contribution_room.remaining) }} available</p>
                        <div v-if="account.contribution_status.length" class="mb-4 text-sm">
                            <p class="m-0 font-medium">Contribution limits</p>
                            <div v-for="status in account.contribution_status" :key="status.year" class="mt-1 flex justify-between text-textSecondary">
                                <span>{{ status.year }}</span>
                                <span>{{ formatMoney(status.contributed) }} of {{ formatMoney(status.annual_limit) }} · {{ formatMoney(status.remaining) }} left</span>
                            </div>
                        </div>
                        <p class="m-0 text-sm font-medium">Deposits</p>
                        <p v-if="!(deposits[account.ID] ?? []).length" class="mb-0 mt-2 text-sm text-textSecondary">No deposits recorded.</p>
                        <div v-for="deposit in deposits[account.ID] ?? []" :key="deposit.ID" class="mt-2 flex justify-between text-sm">
                            <span>{{ deposit.date.slice(0, 10) }}</span>
                            <span>{{ formatMoney(deposit.amount) }}</span>
                        </div>
                        <form class="mt-4 grid gap-2 sm:grid-cols-[1fr_1fr_auto]" @submit.prevent="addDeposit(account.ID)">
                            <input v-model="depositAmounts[account.ID]" required min="0.01" step="0.01" type="number" placeholder="Deposit amount" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                            <input v-model="depositDates[account.ID]" required type="date" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                            <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Add deposit</button>
                        </form>
                    </div>
                </article>
            </section>
        </template>
    </div>
</template>
