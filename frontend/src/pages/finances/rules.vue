<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { ChevronDown, ChevronRight, Pencil, Trash2 } from "lucide-vue-next";
import {
    deleteContributionLimit,
    deleteContributionType,
    saveBirthYear,
    upsertContributionLimit,
} from "~/api/money/api";
import { useContributionTypes, useMoneyMutation } from "~/api/money/queries";
import type { ContributionType } from "~/api/money/types";
import { getApiErrorMessage } from "~/api/httpError";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import ModuleTitleBar, { type ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";
import ContributionTypeDialog from "./components/ContributionTypeDialog.vue";
import PresetDialog from "./components/PresetDialog.vue";
import RoomBreakdown from "./components/RoomBreakdown.vue";
import RoomCard from "./components/RoomCard.vue";
import { formatMoney, parseAmount, ui } from "./format";

const typesQuery = useContributionTypes();
const types = computed(() => typesQuery.data.value?.contribution_types ?? []);
const birthYear = computed(() => typesQuery.data.value?.birth_year ?? null);
const needsBirthYear = computed(
    () => birthYear.value === null && types.value.some((t) => t.eligibility_age !== null),
);

const expanded = ref(new Set<number>());
function toggleBreakdown(id: number) {
    const next = new Set(expanded.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expanded.value = next;
}

// ---- Rule summary ----

function ruleChips(type: ContributionType): string[] {
    const chips: string[] = [];
    if (type.opening_room_year !== null && type.opening_room !== null) {
        chips.push(`Opening room ${formatMoney(type.opening_room)} on Jan 1, ${type.opening_room_year}`);
    } else if (type.room) {
        const why = [];
        if (type.start_year !== null) why.push(`start year ${type.start_year}`);
        if (type.eligibility_age !== null && birthYear.value !== null) why.push(`age ${type.eligibility_age} in ${birthYear.value + type.eligibility_age}`);
        chips.push(`Room since ${type.room.start_year}${why.length ? ` (${why.join(", ")})` : ""}`);
    }
    if (type.carry_forward_cap === null) chips.push("Unused room carries forward");
    else if (type.carry_forward_cap === 0) chips.push("Unused room is lost each year");
    else chips.push(`Carry forward up to ${formatMoney(type.carry_forward_cap)}`);
    if (type.lifetime_limit !== null) chips.push(`Lifetime limit ${formatMoney(type.lifetime_limit)}`);
    chips.push(type.withdrawals_restore_room ? "Withdrawals restored next year" : "Withdrawals don't restore room");
    return chips;
}

function limitRange(type: ContributionType, index: number) {
    // Limits are sorted newest first; each applies until the next newer one.
    const limit = type.limits[index];
    const newer = type.limits[index - 1];
    if (!newer) return `${limit.year} onward`;
    if (newer.year - 1 === limit.year) return `${limit.year}`;
    return `${limit.year}–${newer.year - 1}`;
}

// ---- Mutations ----

const limitInputs = reactive<Record<number, { year: string; amount: string }>>({});
function limitInput(id: number) {
    limitInputs[id] ??= { year: String(new Date().getFullYear()), amount: "" };
    return limitInputs[id];
}

const limitMutation = useMoneyMutation(
    (args: { id: number; year: number; amount: number }) =>
        upsertContributionLimit(args.id, args.year, args.amount),
);
const removeLimitMutation = useMoneyMutation((args: { id: number; year: number }) =>
    deleteContributionLimit(args.id, args.year),
);
const deleteTypeMutation = useMoneyMutation(deleteContributionType);
const birthYearMutation = useMoneyMutation(saveBirthYear);

async function saveLimit(type: ContributionType) {
    const input = limitInput(type.ID);
    const year = parseAmount(input.year);
    const amount = parseAmount(input.amount);
    if (year === null || amount === null) return;
    try {
        await limitMutation.mutateAsync({ id: type.ID, year: Math.trunc(year), amount });
        input.amount = "";
        toast.push(`${type.name} limit for ${year} saved`, "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save limit"), "error");
    }
}

async function removeLimit(type: ContributionType, year: number) {
    const confirmed = await dialogManager.confirm({
        title: "Remove annual limit?",
        message: `The ${type.name} limit starting ${year} will be removed; earlier limits extend to cover those years.`,
        confirmText: "Remove",
    });
    if (!confirmed) return;
    try {
        await removeLimitMutation.mutateAsync({ id: type.ID, year });
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to remove limit"), "error");
    }
}

async function removeType(type: ContributionType) {
    const confirmed = await dialogManager.confirm({
        title: `Delete ${type.name}?`,
        message: type.accounts.length
            ? `Unlink ${type.accounts.map((a) => a.name).join(", ")} first.`
            : "Its rules and limits will be deleted. Accounts and deposits are not affected.",
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteTypeMutation.mutateAsync(type.ID);
        toast.push(`${type.name} deleted`, "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to delete"), "error");
    }
}

function editType(type: ContributionType) {
    void dialogManager.custom<boolean>({
        title: `${type.name} rules`,
        component: ContributionTypeDialog,
        componentProps: { contributionType: type },
    });
}

function openPreset() {
    void dialogManager.custom<boolean>({
        title: "Add registered account type",
        component: PresetDialog,
        componentProps: { existingNames: types.value.map((t) => t.name) },
    });
}

function openCustom() {
    void dialogManager.custom<boolean>({
        title: "Custom account type",
        component: ContributionTypeDialog,
    });
}

const birthYearInput = ref("");
async function submitBirthYear() {
    const year = parseAmount(birthYearInput.value);
    if (year === null) return;
    try {
        await birthYearMutation.mutateAsync(Math.trunc(year));
        toast.push("Birth year saved", "success");
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save birth year"), "error");
    }
}

const navLinks: ModuleTitleLink[] = [
    { name: "Accounts", to: "finances" },
    { name: "Add TFSA / RRSP / FHSA", onClick: openPreset },
    { name: "Custom type", onClick: openCustom },
];
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar title="Contribution room" :links="navLinks" />

        <p v-if="typesQuery.error.value" class="m-0 text-sm text-(--color-cf-red)">
            {{ typesQuery.error.value.message || "Failed to load contribution room" }}
        </p>
        <p v-else-if="typesQuery.isPending.value" class="m-0 text-sm text-textSecondary">Loading…</p>

        <template v-else>
            <form
                v-if="needsBirthYear"
                class="flex flex-wrap items-end gap-3 rounded-lg border border-amber-700/50 bg-amber-950/30 p-4"
                @submit.prevent="submitBirthYear"
            >
                <label :class="[ui.label, 'flex-1 text-amber-100']">
                    Your birth year is needed to know when room started accruing
                    <input v-model="birthYearInput" inputmode="numeric" required placeholder="e.g. 1998" :class="[ui.input, 'max-w-40']" />
                </label>
                <button type="submit" :class="ui.btnPrimary" :disabled="birthYearMutation.isPending.value">Save</button>
            </form>

            <section v-if="types.length === 0" :class="[ui.card, 'flex flex-col items-start gap-3']">
                <h2 class="m-0 text-base font-semibold">No registered accounts yet</h2>
                <p class="m-0 text-sm text-textSecondary">
                    Add a TFSA, RRSP or FHSA with CRA's rules and limits filled in, then link your accounts to it. Every
                    deposit and withdrawal in a linked account counts against the shared room.
                </p>
                <div class="flex flex-wrap gap-2">
                    <button type="button" :class="ui.btnPrimary" @click="openPreset">Add TFSA / RRSP / FHSA</button>
                    <button type="button" :class="ui.btnSecondary" @click="openCustom">Custom type</button>
                </div>
            </section>

            <section v-for="type in types" :key="type.ID" :class="[ui.card, 'flex flex-col gap-4']">
                <header class="flex flex-col gap-2">
                    <div class="flex items-center justify-between gap-3">
                        <h2 class="m-0 min-w-0 truncate text-lg font-semibold text-textPrimary">{{ type.name }}</h2>
                        <div class="flex shrink-0 gap-1">
                            <button type="button" :class="[ui.btnSecondary, 'flex items-center gap-1']" @click="editType(type)">
                                <Pencil :size="14" /> Rules
                            </button>
                            <button type="button" :class="ui.btnDanger" aria-label="Delete type" @click="removeType(type)">
                                <Trash2 :size="16" />
                            </button>
                        </div>
                    </div>
                    <ul class="m-0 flex list-none flex-wrap gap-1.5 p-0">
                        <li v-for="chip in ruleChips(type)" :key="chip" class="rounded bg-secondBg px-2 py-0.5 text-xs text-textSecondary">
                            {{ chip }}
                        </li>
                    </ul>
                </header>

                <div class="grid gap-4 md:grid-cols-[minmax(0,18rem)_1fr]">
                    <RoomCard :contribution-type="type" bare class="self-start" />

                    <div class="flex flex-col gap-3">
                        <div>
                            <h3 class="m-0 text-sm font-semibold text-textPrimary">Annual limits</h3>
                            <p class="m-0 text-xs text-textSecondary">Each limit applies from its year until a newer one replaces it.</p>
                        </div>
                        <p v-if="type.limits.length === 0" class="m-0 text-sm text-textSecondary">
                            No annual limits{{ type.opening_room !== null ? " — only the opening room counts." : "." }}
                        </p>
                        <ul v-else class="m-0 flex list-none flex-col divide-y divide-(--color-border) p-0">
                            <li v-for="(limit, index) in type.limits" :key="limit.ID" class="flex items-center justify-between gap-2 py-1.5 text-sm">
                                <span class="text-textSecondary">{{ limitRange(type, index) }}</span>
                                <span class="flex items-center gap-2">
                                    <span class="tabular-nums text-textPrimary">{{ formatMoney(limit.annual_limit) }} / yr</span>
                                    <button type="button" class="rounded p-1 text-textSecondary hover:bg-secondBg hover:text-(--color-cf-red)" :aria-label="`Remove ${limit.year} limit`" @click="removeLimit(type, limit.year)">
                                        <Trash2 :size="14" />
                                    </button>
                                </span>
                            </li>
                        </ul>
                        <form class="grid grid-cols-[5.5rem_1fr_auto] gap-2" @submit.prevent="saveLimit(type)">
                            <input v-model="limitInput(type.ID).year" inputmode="numeric" required aria-label="Year" :class="ui.input" />
                            <input v-model="limitInput(type.ID).amount" inputmode="decimal" required placeholder="Annual limit" :class="ui.input" />
                            <button type="submit" :class="ui.btnSecondary" :disabled="limitMutation.isPending.value">Set</button>
                        </form>

                        <div v-if="type.accounts.length" class="text-sm">
                            <h3 class="m-0 text-sm font-semibold text-textPrimary">Linked accounts</h3>
                            <div class="mt-1 flex flex-wrap gap-2">
                                <router-link
                                    v-for="account in type.accounts"
                                    :key="account.id"
                                    :to="{ name: 'finance-account', params: { id: account.id } }"
                                    class="rounded bg-secondBg px-2 py-1 text-xs hover:bg-thirdBg"
                                >
                                    {{ account.name }}
                                </router-link>
                            </div>
                        </div>
                        <p v-else class="m-0 text-xs text-textSecondary">
                            No accounts linked — edit an account and choose {{ type.name }} as its registered account.
                        </p>
                    </div>
                </div>

                <div v-if="type.room && type.room.years.length">
                    <button type="button" class="flex items-center gap-1 p-0 text-sm text-textSecondary hover:text-textPrimary" @click="toggleBreakdown(type.ID)">
                        <component :is="expanded.has(type.ID) ? ChevronDown : ChevronRight" :size="16" />
                        Year-by-year breakdown
                    </button>
                    <div v-if="expanded.has(type.ID)" class="mt-3">
                        <RoomBreakdown :room="type.room" />
                    </div>
                </div>
            </section>
        </template>
    </div>
</template>
