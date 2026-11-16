<script setup lang="ts">
import type { RoomSummary } from "~/api/money/types";
import { formatMoney } from "../format";

defineProps<{ room: RoomSummary }>();

function dash(value: number) {
    return value === 0 ? "—" : formatMoney(value);
}
</script>

<template>
    <div class="overflow-x-auto">
        <table class="w-full min-w-[34rem] border-collapse text-right text-xs tabular-nums">
            <thead class="text-textSecondary">
                <tr class="border-b border-(--color-border)">
                    <th class="py-2 pr-2 text-left font-medium">Year</th>
                    <th class="px-2 py-2 font-medium">Carried in</th>
                    <th class="px-2 py-2 font-medium">New room</th>
                    <th class="px-2 py-2 font-medium">Restored</th>
                    <th class="px-2 py-2 font-medium">Available</th>
                    <th class="px-2 py-2 font-medium">Contributed</th>
                    <th class="px-2 py-2 font-medium">Withdrawn</th>
                    <th class="py-2 pl-2 font-medium">Left</th>
                </tr>
            </thead>
            <tbody>
                <tr
                    v-for="year in room.years"
                    :key="year.year"
                    class="border-b border-(--color-border)/50 last:border-0"
                    :class="year.over_contributed ? 'bg-red-950/30' : ''"
                >
                    <td class="py-1.5 pr-2 text-left font-medium text-textPrimary">{{ year.year }}</td>
                    <td class="px-2 py-1.5 text-textSecondary">{{ dash(year.carried_in) }}</td>
                    <td class="px-2 py-1.5">
                        <span v-if="year.is_opening" title="Opening room">{{ formatMoney(year.new_room) }}*</span>
                        <span v-else :class="year.limit_year !== null && year.limit_year !== year.year ? 'text-textSecondary' : ''" :title="year.limit_year !== null && year.limit_year !== year.year ? `Limit from ${year.limit_year}` : ''">
                            {{ dash(year.new_room) }}
                        </span>
                    </td>
                    <td class="px-2 py-1.5 text-textSecondary">{{ dash(year.restored) }}</td>
                    <td class="px-2 py-1.5">
                        {{ formatMoney(year.available) }}
                        <span v-if="year.lifetime_capped > 0" class="block text-[0.65rem] text-amber-300">lifetime cap −{{ formatMoney(year.lifetime_capped) }}</span>
                    </td>
                    <td class="px-2 py-1.5">{{ dash(year.contributed) }}</td>
                    <td class="px-2 py-1.5 text-textSecondary">{{ dash(year.withdrawn) }}</td>
                    <td class="py-1.5 pl-2 font-medium" :class="year.remaining < 0 ? 'text-red-400' : 'text-textPrimary'">
                        {{ formatMoney(year.remaining) }}
                        <span v-if="year.carried_out !== year.remaining" class="block text-[0.65rem] font-normal text-textSecondary">
                            {{ formatMoney(year.carried_out) }} carries over
                        </span>
                    </td>
                </tr>
            </tbody>
        </table>
        <p class="mb-0 mt-2 text-[0.7rem] text-textSecondary">
            Greyed new room is inherited from an earlier year's limit.
            <template v-if="room.years.some((y) => y.is_opening)"> * Opening room you entered.</template>
        </p>
    </div>
</template>
