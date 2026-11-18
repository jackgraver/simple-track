<script setup lang="ts">
import { computed, ref } from "vue";
import Chart from "primevue/chart";
import { AlertTriangle, ChevronRight, Plus, RefreshCw } from "lucide-vue-next";
import {
    useAccounts,
    useContributionTypes,
    useNetWorthHistory,
} from "~/api/money/queries";
import type { Account, AccountKind } from "~/api/money/types";
import { dialogManager } from "~/composables/dialog/useDialog";
import ModuleTitleBar, { type ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";
import AccountFormDialog from "./components/AccountFormDialog.vue";
import UpdateBalancesDialog from "./components/UpdateBalancesDialog.vue";
import RoomCard from "./components/RoomCard.vue";
import { buildMoneyLineChart } from "./chart";
import {
    STALE_BALANCE_DAYS,
    accountKindLabels,
    dayKey,
    daysSince,
    formatAge,
    formatMoney,
    formatSignedMoney,
    ui,
} from "./format";
import { useRouter } from "vue-router";

const router = useRouter();
const showArchived = ref(false);

const accountsQuery = useAccounts(showArchived);
const historyQuery = useNetWorthHistory();
const typesQuery = useContributionTypes();

const accounts = computed(() => accountsQuery.data.value ?? []);
const activeAccounts = computed(() => accounts.value.filter((a) => !a.archived_at));
const archivedAccounts = computed(() => accounts.value.filter((a) => a.archived_at));
const contributionTypes = computed(() => typesQuery.data.value?.contribution_types ?? []);

const assets = computed(() =>
    activeAccounts.value.filter((a) => !a.is_liability).reduce((sum, a) => sum + a.current_balance, 0),
);
const liabilities = computed(() =>
    activeAccounts.value.filter((a) => a.is_liability).reduce((sum, a) => sum + a.current_balance, 0),
);
const netWorth = computed(() => assets.value - liabilities.value);

const staleAccounts = computed(() =>
    activeAccounts.value.filter(
        (a) => !a.balance_date || daysSince(a.balance_date) > STALE_BALANCE_DAYS,
    ),
);

// ---- Net worth history ----

type Range = "3m" | "1y" | "all";
const range = ref<Range>("1y");
const rangeOptions: { value: Range; label: string }[] = [
    { value: "3m", label: "3M" },
    { value: "1y", label: "1Y" },
    { value: "all", label: "All" },
];

function cutoffFor(value: Range) {
    if (value === "all") return "";
    const d = new Date();
    d.setMonth(d.getMonth() - (value === "3m" ? 3 : 12));
    return d.toISOString().slice(0, 10);
}

const history = computed(() => historyQuery.data.value ?? []);
const visibleHistory = computed(() => {
    const cutoff = cutoffFor(range.value);
    const inRange = history.value.filter((p) => dayKey(p.date) >= cutoff);
    // Anchor the start of the range with the last point before it so the
    // line doesn't start mid-air.
    const earlier = history.value.filter((p) => dayKey(p.date) < cutoff);
    const before = earlier[earlier.length - 1];
    return before ? [before, ...inRange] : inRange;
});
const rangeChange = computed(() => {
    const points = visibleHistory.value;
    if (points.length < 2) return null;
    return points[points.length - 1].net_worth - points[0].net_worth;
});
const chart = computed(() =>
    buildMoneyLineChart(
        visibleHistory.value.map((p) => ({ date: p.date, value: p.net_worth })),
        "Net worth",
    ),
);

// ---- Account groups ----

const groups: { title: string; kinds: AccountKind[] }[] = [
    { title: "Cash & banking", kinds: ["chequing", "savings", "cash"] },
    { title: "Investments", kinds: ["investment"] },
    { title: "Property & other", kinds: ["property", "other"] },
    { title: "Liabilities", kinds: ["credit_card", "loan"] },
];

const groupedAccounts = computed(() =>
    groups
        .map((group) => {
            const list = activeAccounts.value.filter((a) => group.kinds.includes(a.kind));
            return {
                ...group,
                accounts: list,
                total: list.reduce((sum, a) => sum + a.current_balance, 0),
            };
        })
        .filter((group) => group.accounts.length > 0),
);

function balanceChange(account: Account) {
    if (account.previous_balance === null) return null;
    const diff = Math.round((account.current_balance - account.previous_balance) * 100) / 100;
    return diff === 0 ? null : diff;
}

function changeClass(account: Account, diff: number) {
    return diff > 0 !== account.is_liability ? ui.positive : ui.negative;
}

function isStale(account: Account) {
    return !account.balance_date || daysSince(account.balance_date) > STALE_BALANCE_DAYS;
}

// ---- Actions ----

async function openAddAccount() {
    const id = await dialogManager.custom<number>({
        title: "Add account",
        component: AccountFormDialog,
        componentProps: { contributionTypes: contributionTypes.value },
    });
    if (typeof id === "number") {
        void router.push({ name: "finance-account", params: { id } });
    }
}

function openUpdateBalances() {
    void dialogManager.custom<boolean>({
        title: "Update balances",
        component: UpdateBalancesDialog,
        componentProps: { accounts: activeAccounts.value },
    });
}

const navLinks = computed<ModuleTitleLink[]>(() => [
    { name: "Update balances", onClick: openUpdateBalances },
    { name: "Add account", onClick: openAddAccount },
]);

const isLoading = computed(() => accountsQuery.isPending.value);
const loadError = computed(() => accountsQuery.error.value);
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar title="Finances" :links="navLinks" />

        <p v-if="loadError" class="m-0 text-sm text-(--color-cf-red)">
            {{ loadError.message || "Failed to load accounts" }}
        </p>
        <p v-else-if="isLoading" class="m-0 text-sm text-textSecondary">Loading accounts…</p>

        <section v-else-if="activeAccounts.length === 0 && !showArchived" :class="[ui.card, 'flex flex-col items-start gap-3']">
            <h2 class="m-0 text-base font-semibold text-textPrimary">Track every account in one place</h2>
            <p class="m-0 text-sm text-textSecondary">
                Add your chequing, savings, investments, credit cards and loans. Update balances whenever you check them and
                you'll get a net worth history over time. Link TFSA / RRSP / FHSA accounts to see your contribution room.
            </p>
            <div class="flex flex-wrap gap-2">
                <button type="button" :class="ui.btnPrimary" @click="openAddAccount">Add your first account</button>
                <router-link :to="{ name: 'finance-rules' }" :class="ui.btnSecondary">Set up contribution room</router-link>
                <button type="button" class="px-1 text-sm text-textSecondary hover:text-textPrimary" @click="showArchived = true">
                    Show archived
                </button>
            </div>
        </section>

        <template v-else>
            <!-- Net worth -->
            <section :class="[ui.card, 'flex flex-col gap-4']">
                <div class="flex flex-wrap items-end justify-between gap-4">
                    <div>
                        <p class="m-0 text-sm text-textSecondary">Net worth</p>
                        <p class="m-0 text-3xl font-semibold tabular-nums text-textPrimary">{{ formatMoney(netWorth) }}</p>
                        <p v-if="rangeChange !== null" class="m-0 mt-1 text-sm tabular-nums" :class="rangeChange >= 0 ? ui.positive : ui.negative">
                            {{ formatSignedMoney(rangeChange) }}
                            <span class="text-textSecondary">over {{ range === "all" ? "all time" : range === "1y" ? "the past year" : "3 months" }}</span>
                        </p>
                    </div>
                    <dl class="m-0 flex gap-6 text-sm">
                        <div>
                            <dt class="text-textSecondary">Assets</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(assets) }}</dd>
                        </div>
                        <div>
                            <dt class="text-textSecondary">Liabilities</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(liabilities) }}</dd>
                        </div>
                    </dl>
                </div>
                <div v-if="history.length > 1" class="flex flex-col gap-2">
                    <div class="flex justify-end gap-1">
                        <button
                            v-for="option in rangeOptions"
                            :key="option.value"
                            type="button"
                            class="rounded px-2 py-0.5 text-xs transition-colors"
                            :class="range === option.value ? 'bg-secondBg text-textPrimary' : 'text-textSecondary hover:text-textPrimary'"
                            @click="range = option.value"
                        >
                            {{ option.label }}
                        </button>
                    </div>
                    <div class="h-56 w-full min-w-0">
                        <Chart type="line" :data="chart.data" :options="chart.options" class="h-full w-full" />
                    </div>
                </div>
                <p v-else class="m-0 text-xs text-textSecondary">
                    Record balances on different days to build a net worth history.
                </p>
            </section>

            <!-- Stale balances -->
            <button
                v-if="staleAccounts.length"
                type="button"
                class="flex items-center gap-3 rounded-lg border border-amber-700/50 bg-amber-950/30 px-4 py-3 text-left text-sm text-amber-100 transition-colors hover:bg-amber-950/50"
                @click="openUpdateBalances"
            >
                <AlertTriangle :size="18" class="shrink-0" />
                <span class="flex-1">
                    {{ staleAccounts.length }} balance{{ staleAccounts.length === 1 ? " hasn't" : "s haven't" }} been updated in over a month
                </span>
                <span class="flex items-center gap-1 font-medium">Update <RefreshCw :size="14" /></span>
            </button>

            <!-- Contribution room -->
            <section v-if="contributionTypes.length" class="flex flex-col gap-3">
                <div class="flex items-center justify-between">
                    <h2 class="m-0 text-base font-semibold text-textPrimary">Contribution room</h2>
                    <router-link :to="{ name: 'finance-rules' }" class="flex items-center text-sm text-textSecondary hover:text-textPrimary">
                        Details <ChevronRight :size="16" />
                    </router-link>
                </div>
                <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    <RoomCard v-for="type in contributionTypes" :key="type.ID" :contribution-type="type" compact />
                </div>
            </section>

            <!-- Accounts -->
            <section v-for="group in groupedAccounts" :key="group.title" class="flex flex-col gap-2">
                <div class="flex items-baseline justify-between">
                    <h2 class="m-0 text-base font-semibold text-textPrimary">{{ group.title }}</h2>
                    <span class="text-sm tabular-nums text-textSecondary">{{ formatMoney(group.total) }}</span>
                </div>
                <div class="flex flex-col divide-y divide-(--color-border) overflow-hidden rounded-lg border border-(--color-border) bg-firstBg">
                    <router-link
                        v-for="account in group.accounts"
                        :key="account.ID"
                        :to="{ name: 'finance-account', params: { id: account.ID } }"
                        class="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-secondBg"
                    >
                        <div class="min-w-0 flex-1">
                            <p class="m-0 flex items-center gap-2 truncate font-medium text-textPrimary">
                                {{ account.name }}
                                <span v-if="account.contribution_type" class="rounded bg-secondBg px-1.5 py-0.5 text-[0.65rem] font-semibold uppercase tracking-wide text-textSecondary">
                                    {{ account.contribution_type.name }}
                                </span>
                            </p>
                            <p class="m-0 truncate text-xs text-textSecondary">
                                {{ account.institution || accountKindLabels[account.kind] }}
                                ·
                                <span :class="isStale(account) ? 'text-amber-400' : ''">
                                    {{ account.balance_date ? `updated ${formatAge(account.balance_date)}` : "no balance yet" }}
                                </span>
                            </p>
                        </div>
                        <div class="text-right">
                            <p class="m-0 font-medium tabular-nums text-textPrimary">{{ formatMoney(account.current_balance) }}</p>
                            <p v-if="balanceChange(account) !== null" class="m-0 text-xs tabular-nums" :class="changeClass(account, balanceChange(account)!)">
                                {{ formatSignedMoney(balanceChange(account)!) }}
                            </p>
                        </div>
                        <ChevronRight :size="16" class="shrink-0 text-textSecondary" />
                    </router-link>
                </div>
            </section>

            <section v-if="showArchived && archivedAccounts.length" class="flex flex-col gap-2">
                <h2 class="m-0 text-base font-semibold text-textSecondary">Archived</h2>
                <div class="flex flex-col divide-y divide-(--color-border) overflow-hidden rounded-lg border border-(--color-border) bg-firstBg opacity-70">
                    <router-link
                        v-for="account in archivedAccounts"
                        :key="account.ID"
                        :to="{ name: 'finance-account', params: { id: account.ID } }"
                        class="flex items-center justify-between gap-3 px-4 py-3 hover:bg-secondBg"
                    >
                        <span class="truncate">{{ account.name }}</span>
                        <span class="text-sm tabular-nums text-textSecondary">{{ formatMoney(account.current_balance) }}</span>
                    </router-link>
                </div>
            </section>

            <div class="flex flex-wrap items-center justify-between gap-2">
                <button type="button" class="text-sm text-textSecondary hover:text-textPrimary" @click="showArchived = !showArchived">
                    {{ showArchived ? "Hide archived accounts" : "Show archived accounts" }}
                </button>
                <button type="button" :class="[ui.btnSecondary, 'flex items-center gap-1']" @click="openAddAccount">
                    <Plus :size="16" /> Add account
                </button>
            </div>
        </template>
    </div>
</template>
