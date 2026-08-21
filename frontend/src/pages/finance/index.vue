<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import {
    createInvestmentAccount,
    createInvestmentAccountType,
    createInvestmentDeposit,
    deleteContributionRule,
    deleteInvestmentAccount,
    deleteInvestmentAccountType,
    deleteInvestmentDeposit,
    listInvestmentAccounts,
    listInvestmentAccountTypes,
    listInvestmentDeposits,
    updateInvestmentAccount,
    upsertContributionRule,
    type InvestmentAccount,
    type InvestmentAccountType,
    type InvestmentDeposit,
} from "~/api/money/api";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

const accounts = ref<InvestmentAccount[]>([]);
const accountTypes = ref<InvestmentAccountType[]>([]);
const deposits = ref<Record<number, InvestmentDeposit[]>>({});
const isLoading = ref(true);
const isSaving = ref(false);
const errorMessage = ref("");
const accountTypeName = ref("");
const contributionStartAge = ref("");
const accountName = ref("");
const accountTypeId = ref("");
const currentBalance = ref("");
const depositAmounts = reactive<Record<number, string>>({});
const depositDates = reactive<Record<number, string>>({});
const balanceInputs = reactive<Record<number, string>>({});
const ruleYears = reactive<Record<number, string>>({});
const ruleLimits = reactive<Record<number, string>>({});

function today() {
    return new Date().toISOString().slice(0, 10);
}

function formatMoney(value: number) {
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
    }).format(value);
}

async function load() {
    isLoading.value = true;
    errorMessage.value = "";
    try {
        const [nextAccounts, nextAccountTypes] = await Promise.all([
            listInvestmentAccounts(),
            listInvestmentAccountTypes(),
        ]);
        accounts.value = nextAccounts;
        accountTypes.value = nextAccountTypes;
        const accountDeposits = await Promise.all(
            nextAccounts.map(async (account) => [
                account.ID,
                await listInvestmentDeposits(account.ID),
            ] as const),
        );
        deposits.value = Object.fromEntries(accountDeposits);
        for (const account of nextAccounts) {
            depositDates[account.ID] ??= today();
            balanceInputs[account.ID] = String(account.current_balance);
        }
    } catch (error) {
        errorMessage.value = getApiErrorMessage(error, "Failed to load investments");
    } finally {
        isLoading.value = false;
    }
}

async function run(action: () => Promise<void>, successMessage: string) {
    if (isSaving.value) return;
    isSaving.value = true;
    try {
        await action();
        await load();
        toast.push(successMessage, "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Something went wrong"), "error");
    } finally {
        isSaving.value = false;
    }
}

function addAccountType() {
    const name = accountTypeName.value.trim();
    if (!name) return;
    void run(async () => {
        await createInvestmentAccountType({
            name,
            contribution_start_age: contributionStartAge.value
                ? Number(contributionStartAge.value)
                : undefined,
        });
        accountTypeName.value = "";
        contributionStartAge.value = "";
    }, "Account type created");
}

function addAccount() {
    const name = accountName.value.trim();
    if (!name || currentBalance.value === "") return;
    void run(async () => {
        await createInvestmentAccount({
            name,
            current_balance: Number(currentBalance.value),
            investment_account_type_id: accountTypeId.value
                ? Number(accountTypeId.value)
                : undefined,
        });
        accountName.value = "";
        accountTypeId.value = "";
        currentBalance.value = "";
    }, "Investment account created");
}

function addDeposit(accountId: number) {
    const amount = depositAmounts[accountId];
    const date = depositDates[accountId];
    if (!amount || !date) return;
    void run(async () => {
        await createInvestmentDeposit(accountId, {
            amount: Number(amount),
            date,
        });
        depositAmounts[accountId] = "";
    }, "Deposit added");
}

function saveBalance(account: InvestmentAccount) {
    const balance = balanceInputs[account.ID];
    if (balance === undefined || balance === "") return;
    void run(
        () =>
            updateInvestmentAccount(account.ID, {
                current_balance: Number(balance),
            }),
        "Balance updated",
    );
}

function saveRule(accountTypeId: number) {
    const year = ruleYears[accountTypeId];
    const limit = ruleLimits[accountTypeId];
    if (!year || limit === undefined || limit === "") return;
    void run(async () => {
        await upsertContributionRule(accountTypeId, Number(year), Number(limit));
        ruleYears[accountTypeId] = "";
        ruleLimits[accountTypeId] = "";
    }, "Contribution limit saved");
}

function removeAccount(account: InvestmentAccount) {
    if (!window.confirm(`Delete ${account.name}?`)) return;
    void run(
        () => deleteInvestmentAccount(account.ID),
        "Investment account deleted",
    );
}

function removeAccountType(accountType: InvestmentAccountType) {
    if (!window.confirm(`Delete ${accountType.name}?`)) return;
    void run(
        () => deleteInvestmentAccountType(accountType.ID),
        "Account type deleted",
    );
}

function removeDeposit(accountId: number, depositId: number) {
    if (!window.confirm("Delete this deposit?")) return;
    void run(
        () => deleteInvestmentDeposit(accountId, depositId),
        "Deposit deleted",
    );
}

function removeRule(accountTypeId: number, year: number) {
    if (!window.confirm(`Delete the ${year} contribution limit?`)) return;
    void run(
        () => deleteContributionRule(accountTypeId, year),
        "Contribution limit deleted",
    );
}

onMounted(() => {
    void load();
});
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-8">
        <div class="flex items-center justify-between gap-4">
            <div>
                <h1 class="m-0 text-xl font-semibold text-textPrimary">Finance</h1>
                <p class="mb-0 mt-1 text-sm text-textSecondary">Investment accounts and contributions</p>
            </div>
            <button
                type="button"
                class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg disabled:opacity-50"
                :disabled="isLoading || isSaving"
                @click="load"
            >Refresh</button>
        </div>
        <p v-if="errorMessage" class="m-0 text-sm text-(--color-cf-red)">{{ errorMessage }}</p>
        <p v-else-if="isLoading" class="m-0 text-sm text-textSecondary">Loading investments…</p>
        <template v-else>
            <section class="rounded-lg bg-secondBg p-4">
                <h2 class="m-0 text-base font-semibold">Add investment account</h2>
                <form class="mt-4 grid gap-3 md:grid-cols-4" @submit.prevent="addAccount">
                    <input v-model="accountName" required placeholder="Account name" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <select v-model="accountTypeId" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none">
                        <option value="">No account type</option>
                        <option v-for="type in accountTypes" :key="type.ID" :value="type.ID">{{ type.name }}</option>
                    </select>
                    <input v-model="currentBalance" required min="0" step="0.01" type="number" placeholder="Current balance" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Add account</button>
                </form>
            </section>
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold">Accounts</h2>
                <p v-if="accounts.length === 0" class="m-0 text-sm text-textSecondary">No investment accounts yet.</p>
                <article v-for="account in accounts" :key="account.ID" class="rounded-lg bg-secondBg p-4">
                    <div class="flex flex-wrap items-start justify-between gap-3">
                        <div>
                            <h3 class="m-0 font-medium">{{ account.name }}</h3>
                            <p class="mb-0 mt-1 text-sm text-textSecondary">{{ account.investment_account_type?.name ?? "Uncategorized" }}</p>
                        </div>
                        <button type="button" class="text-sm text-(--color-cf-red) hover:underline" :disabled="isSaving" @click="removeAccount(account)">Delete</button>
                    </div>
                    <dl class="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
                        <div><dt class="text-textSecondary">Balance</dt><dd class="m-0 font-medium">{{ formatMoney(account.current_balance) }}</dd></div>
                        <div><dt class="text-textSecondary">Deposits</dt><dd class="m-0 font-medium">{{ formatMoney(account.total_deposits) }}</dd></div>
                        <div><dt class="text-textSecondary">Profit</dt><dd class="m-0 font-medium" :class="account.profit >= 0 ? 'text-emerald-400' : 'text-(--color-cf-red)'">{{ formatMoney(account.profit) }}</dd></div>
                        <div v-if="account.contribution_room"><dt class="text-textSecondary">Contribution room</dt><dd class="m-0 font-medium">{{ formatMoney(account.contribution_room.remaining) }}</dd></div>
                    </dl>
                    <form class="mt-5 grid gap-2 sm:grid-cols-[1fr_auto]" @submit.prevent="saveBalance(account)">
                        <input v-model="balanceInputs[account.ID]" required min="0" step="0.01" type="number" aria-label="Current balance" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <button type="submit" class="rounded-md bg-thirdBg px-3 py-2 text-sm hover:bg-fourthBg disabled:opacity-50" :disabled="isSaving">Update balance</button>
                    </form>
                    <form class="mt-5 grid gap-2 sm:grid-cols-[1fr_1fr_auto]" @submit.prevent="addDeposit(account.ID)">
                        <input v-model="depositAmounts[account.ID]" required min="0.01" step="0.01" type="number" placeholder="Deposit amount" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <input v-model="depositDates[account.ID]" required type="date" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Add deposit</button>
                    </form>
                    <div v-if="(deposits[account.ID] ?? []).length" class="mt-4 border-t border-white/10 pt-3">
                        <div v-for="deposit in deposits[account.ID]" :key="deposit.ID" class="flex items-center justify-between py-1 text-sm">
                            <span>{{ deposit.date.slice(0, 10) }} · {{ formatMoney(deposit.amount) }}</span>
                            <button type="button" class="text-textSecondary hover:text-(--color-cf-red)" :disabled="isSaving" @click="removeDeposit(account.ID, deposit.ID)">Remove</button>
                        </div>
                    </div>
                </article>
            </section>
            <section class="rounded-lg bg-secondBg p-4">
                <h2 class="m-0 text-base font-semibold">Account types</h2>
                <form class="mt-4 grid gap-3 md:grid-cols-3" @submit.prevent="addAccountType">
                    <input v-model="accountTypeName" required placeholder="Type name (e.g. TFSA)" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <input v-model="contributionStartAge" min="0" type="number" placeholder="Contribution start age" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Add type</button>
                </form>
                <p v-if="accountTypes.length === 0" class="mb-0 mt-4 text-sm text-textSecondary">Add an account type to track contribution limits.</p>
                <article v-for="type in accountTypes" :key="type.ID" class="mt-5 border-t border-white/10 pt-4">
                    <div class="flex items-start justify-between gap-3">
                        <div>
                            <h3 class="m-0 font-medium">{{ type.name }}</h3>
                            <p v-if="type.contribution_room" class="mb-0 mt-1 text-sm text-textSecondary">Remaining contribution room: {{ formatMoney(type.contribution_room.remaining) }}</p>
                        </div>
                        <button type="button" class="text-sm text-(--color-cf-red) hover:underline" :disabled="isSaving" @click="removeAccountType(type)">Delete</button>
                    </div>
                    <div v-if="(type.rules ?? []).length" class="mt-3 text-sm">
                        <div v-for="rule in type.rules ?? []" :key="rule.ID" class="flex items-center justify-between py-1">
                            <span>{{ rule.year }} limit: {{ formatMoney(rule.annual_limit) }}</span>
                            <button type="button" class="text-textSecondary hover:text-(--color-cf-red)" :disabled="isSaving" @click="removeRule(type.ID, rule.year)">Remove</button>
                        </div>
                    </div>
                    <form class="mt-3 grid gap-2 sm:grid-cols-[1fr_1fr_auto]" @submit.prevent="saveRule(type.ID)">
                        <input v-model="ruleYears[type.ID]" required min="1900" max="9999" type="number" placeholder="Year" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <input v-model="ruleLimits[type.ID]" required min="0" step="0.01" type="number" placeholder="Annual limit" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <button type="submit" class="rounded-md bg-thirdBg px-3 py-2 text-sm hover:bg-fourthBg disabled:opacity-50" :disabled="isSaving">Save limit</button>
                    </form>
                </article>
            </section>
        </template>
    </div>
</template>
