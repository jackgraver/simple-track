<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import {
    deleteContributionRule,
    deleteInvestmentAccount,
    deleteInvestmentAccountType,
    listInvestmentAccounts,
    listInvestmentAccountTypes,
    updateInvestmentAccount,
    updateInvestmentAccountType,
    upsertContributionRule,
    type InvestmentAccount,
    type InvestmentAccountType,
} from "~/api/money/api";
import { getApiErrorMessage } from "~/api/httpError";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import CreateAccountDialog from "./components/CreateAccountDialog.vue";
import CreateAccountTypeDialog from "./components/CreateAccountTypeDialog.vue";

const accounts = ref<InvestmentAccount[]>([]);
const accountTypes = ref<InvestmentAccountType[]>([]);
const balanceInputs = reactive<Record<number, string>>({});
const startYearInputs = reactive<Record<number, string>>({});
const ruleYears = reactive<Record<number, string>>({});
const ruleLimits = reactive<Record<number, string>>({});
const isLoading = ref(true);
const isSaving = ref(false);
const errorMessage = ref("");

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
        for (const account of nextAccounts) {
            balanceInputs[account.ID] = String(account.current_balance);
        }
        for (const accountType of nextAccountTypes) {
            startYearInputs[accountType.ID] = accountType.contribution_start_year
                ? String(accountType.contribution_start_year)
                : "";
        }
    } catch (error) {
        errorMessage.value = getApiErrorMessage(error, "Failed to load configuration");
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

async function openCreateAccount() {
    const created = await dialogManager.custom<boolean>({
        title: "Add investment account",
        component: CreateAccountDialog,
        componentProps: { accountTypes: accountTypes.value },
    });
    if (created) await load();
}

async function openCreateAccountType() {
    const created = await dialogManager.custom<boolean>({
        title: "Add account type",
        component: CreateAccountTypeDialog,
    });
    if (created) await load();
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

function saveStartYear(accountType: InvestmentAccountType) {
    const startYear = startYearInputs[accountType.ID];
    if (!startYear) return;
    void run(
        () =>
            updateInvestmentAccountType(accountType.ID, {
                name: accountType.name,
                contribution_start_year: Number(startYear),
            }),
        "Contribution start year updated",
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

async function removeAccount(account: InvestmentAccount) {
    const confirmed = await dialogManager.confirm({
        title: "Delete investment account?",
        message: `${account.name} and its deposits will be permanently deleted.`,
        confirmText: "Delete",
    });
    if (!confirmed) return;
    void run(
        () => deleteInvestmentAccount(account.ID),
        "Investment account deleted",
    );
}

async function removeAccountType(accountType: InvestmentAccountType) {
    const confirmed = await dialogManager.confirm({
        title: "Delete account type?",
        message: `${accountType.name} can only be deleted when no accounts use it.`,
        confirmText: "Delete",
    });
    if (!confirmed) return;
    void run(
        () => deleteInvestmentAccountType(accountType.ID),
        "Account type deleted",
    );
}

async function removeRule(accountTypeId: number, year: number) {
    const confirmed = await dialogManager.confirm({
        title: "Delete contribution limit?",
        message: `The ${year} annual limit will be removed.`,
        confirmText: "Delete",
    });
    if (!confirmed) return;
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
        <header class="flex flex-wrap items-start justify-between gap-4">
            <div>
                <h1 class="m-0 text-xl font-semibold text-textPrimary">Finance configuration</h1>
                <p class="mb-0 mt-1 text-sm text-textSecondary">Manage accounts, account types, and contribution limits</p>
            </div>
            <div class="flex gap-2">
                <router-link :to="{ name: 'finances' }" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg">Summary</router-link>
                <button type="button" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg disabled:opacity-50" :disabled="isLoading" @click="load">Refresh</button>
            </div>
        </header>
        <p v-if="errorMessage" class="m-0 text-sm text-(--color-cf-red)">{{ errorMessage }}</p>
        <p v-else-if="isLoading" class="m-0 text-sm text-textSecondary">Loading configuration…</p>
        <template v-else>
            <section class="rounded-lg bg-secondBg p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                        <h2 class="m-0 text-base font-semibold">Investment accounts</h2>
                        <p class="mb-0 mt-1 text-sm text-textSecondary">Set balances here; deposits are managed from account details.</p>
                    </div>
                    <button type="button" class="rounded-md bg-primary px-3 py-2 text-sm text-white" @click="openCreateAccount">Add account</button>
                </div>
                <p v-if="accounts.length === 0" class="mb-0 mt-4 text-sm text-textSecondary">No accounts yet.</p>
                <article v-for="account in accounts" :key="account.ID" class="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-white/10 pt-4">
                    <div>
                        <p class="m-0 font-medium">{{ account.name }}</p>
                        <p class="mb-0 mt-1 text-sm text-textSecondary">{{ account.investment_account_type?.name ?? "Uncategorized" }}</p>
                    </div>
                    <form class="flex items-center gap-2" @submit.prevent="saveBalance(account)">
                        <input v-model="balanceInputs[account.ID]" required min="0" step="0.01" type="number" aria-label="Current balance" class="w-35 rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <button type="submit" class="rounded-md bg-thirdBg px-3 py-2 text-sm hover:bg-fourthBg disabled:opacity-50" :disabled="isSaving">Save</button>
                        <button type="button" class="text-sm text-(--color-cf-red) hover:underline" :disabled="isSaving" @click="removeAccount(account)">Delete</button>
                    </form>
                </article>
            </section>
            <section class="rounded-lg bg-secondBg p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                        <h2 class="m-0 text-base font-semibold">Account types</h2>
                        <p class="mb-0 mt-1 text-sm text-textSecondary">Each annual limit applies from its year until the next rule.</p>
                    </div>
                    <button type="button" class="rounded-md bg-primary px-3 py-2 text-sm text-white" @click="openCreateAccountType">Add account type</button>
                </div>
                <p v-if="accountTypes.length === 0" class="mb-0 mt-4 text-sm text-textSecondary">No account types yet.</p>
                <article v-for="type in accountTypes" :key="type.ID" class="mt-5 border-t border-white/10 pt-4">
                    <div class="flex flex-wrap items-start justify-between gap-3">
                        <div>
                            <h3 class="m-0 font-medium">{{ type.name }}</h3>
                        </div>
                        <button type="button" class="text-sm text-(--color-cf-red) hover:underline" :disabled="isSaving" @click="removeAccountType(type)">Delete</button>
                    </div>
                    <form class="mt-3 grid gap-2 sm:grid-cols-[1fr_auto]" @submit.prevent="saveStartYear(type)">
                        <input v-model="startYearInputs[type.ID]" required min="1900" max="9999" type="number" placeholder="Contribution start year" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                        <button type="submit" class="rounded-md bg-thirdBg px-3 py-2 text-sm hover:bg-fourthBg disabled:opacity-50" :disabled="isSaving">Save start year</button>
                    </form>
                    <div v-if="(type.rules ?? []).length" class="mt-3 text-sm">
                        <div v-for="rule in type.rules ?? []" :key="rule.ID" class="flex items-center justify-between py-1">
                            <span>{{ rule.year }} onward: {{ formatMoney(rule.annual_limit) }} / year</span>
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
