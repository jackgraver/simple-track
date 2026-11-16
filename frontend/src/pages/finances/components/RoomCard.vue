<script setup lang="ts">
import { computed } from "vue";
import { AlertTriangle } from "lucide-vue-next";
import type { ContributionType } from "~/api/money/types";
import { formatMoney, ui } from "../format";

const props = defineProps<{
    contributionType: ContributionType;
    compact?: boolean;
    /** Hide the name/accounts header when the page already shows them. */
    bare?: boolean;
}>();

const room = computed(() => props.contributionType.room);
const currentYear = computed(() => room.value?.years[0]);

/** Share of this year's room already used. */
const usedPercent = computed(() => {
    const year = currentYear.value;
    if (!year || year.available <= 0) return year && year.contributed > 0 ? 100 : 0;
    return Math.min(100, Math.max(0, (year.contributed / year.available) * 100));
});
</script>

<template>
    <article :class="[ui.card, 'flex flex-col gap-3']">
        <div v-if="!bare || room?.over_contributed" class="flex items-start justify-between gap-2">
            <div v-if="!bare" class="min-w-0">
                <h3 class="m-0 text-sm font-semibold text-textPrimary">{{ contributionType.name }}</h3>
                <p class="m-0 truncate text-xs text-textSecondary">
                    {{
                        contributionType.accounts.length
                            ? contributionType.accounts.map((a) => a.name).join(", ")
                            : "No accounts linked"
                    }}
                </p>
            </div>
            <span
                v-if="room?.over_contributed"
                class="flex shrink-0 items-center gap-1 rounded bg-red-950/60 px-1.5 py-0.5 text-xs font-medium text-red-300"
            >
                <AlertTriangle :size="12" /> Over
            </span>
        </div>

        <template v-if="room && currentYear">
            <div>
                <p class="m-0 text-xs text-textSecondary">Available to contribute</p>
                <p class="m-0 text-2xl font-semibold tabular-nums" :class="room.available < 0 ? ui.negative : 'text-textPrimary'">
                    {{ formatMoney(room.available) }}
                </p>
            </div>
            <div class="flex flex-col gap-1">
                <div class="h-1.5 w-full overflow-hidden rounded-full bg-secondBg">
                    <div
                        class="h-full rounded-full"
                        :class="room.over_contributed ? 'bg-red-500' : 'bg-emerald-500'"
                        :style="{ width: `${usedPercent}%` }"
                    />
                </div>
                <p class="m-0 text-xs text-textSecondary">
                    {{ formatMoney(currentYear.contributed) }} of {{ formatMoney(currentYear.available) }} used in {{ currentYear.year }}
                </p>
            </div>
            <dl v-if="!compact || room.lifetime_remaining !== null" class="m-0 grid grid-cols-2 gap-2 text-xs">
                <div v-if="room.lifetime_remaining !== null">
                    <dt class="text-textSecondary">Lifetime left</dt>
                    <dd class="m-0 font-medium tabular-nums">{{ formatMoney(room.lifetime_remaining) }}</dd>
                </div>
                <template v-if="!compact">
                    <div>
                        <dt class="text-textSecondary">New room {{ currentYear.year }}</dt>
                        <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.new_room) }}</dd>
                    </div>
                    <div v-if="currentYear.restored > 0">
                        <dt class="text-textSecondary">Restored withdrawals</dt>
                        <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.restored) }}</dd>
                    </div>
                    <div v-if="currentYear.withdrawn > 0">
                        <dt class="text-textSecondary">Withdrawn {{ currentYear.year }}</dt>
                        <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.withdrawn) }}</dd>
                    </div>
                </template>
            </dl>
        </template>
        <p v-else class="m-0 text-xs text-amber-300">
            {{ contributionType.room_issue || "Room can't be calculated yet." }}
        </p>
    </article>
</template>
