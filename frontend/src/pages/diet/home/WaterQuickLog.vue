<script setup lang="ts">
import { useRouter } from "vue-router";
import { List as ListIcon, Plus } from "lucide-vue-next";
import { newClientId } from "~/api/ids";
import { useDrinkSizePresets, useSaveWater } from "~/api/tracking/queries";
import type { DrinkSizePreset } from "~/api/tracking/types";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { Button, IconButton, Skeleton, useDelayedFlag } from "~/shared/ui";
import LogWaterDialog from "../components/dialog/LogWaterDialog.vue";
import WaterLogsDialog from "../components/dialog/WaterLogsDialog.vue";

const props = defineProps<{ dateStr: string; dayLabel: string }>();

const router = useRouter();
const { formatVolumeFromOz } = useWaterPrefs();
const vol = (oz: number) => {
    const v = formatVolumeFromOz(oz);
    return `${v.value} ${v.unit}`;
};
const { data: presets, isPending, isError } = useDrinkSizePresets();
const showSkeleton = useDelayedFlag(isPending);
const save = useSaveWater();

/** Optimistic: the entry is in the cache before the request returns; a failure rolls back with a Retry toast. */
function logPreset(p: DrinkSizePreset) {
    save.mutate({ client_id: newClientId(), date: props.dateStr, amountOz: p.amount_oz, presetId: p.ID, preset: p });
    toast.push(`Logged ${vol(p.amount_oz)} water`, "success");
}

function openCustom() {
    void dialogManager.custom<boolean>({
        title: "Log water",
        component: LogWaterDialog,
        componentProps: { dateStr: props.dateStr },
    });
}

function openEntries() {
    void dialogManager.custom<void>({
        title: "Water log",
        component: WaterLogsDialog,
        componentProps: { dateStr: props.dateStr, dayLabel: props.dayLabel },
    });
}
</script>

<template>
    <div class="flex flex-wrap items-center gap-2">
        <template v-if="isPending">
            <Skeleton v-for="n in showSkeleton ? 3 : 0" :key="n" shape="block" width="5rem" height="2.75rem" />
        </template>
        <template v-else-if="!isError && (presets ?? []).length">
            <Button
                v-for="p in presets"
                :key="p.ID"
                variant="secondary"
                size="sm"
                :aria-label="`Log ${p.name}, ${vol(p.amount_oz)} of water`"
                @click="logPreset(p)"
            >
                <template #icon-left><Plus class="size-4" /></template>
                {{ vol(p.amount_oz) }}
            </Button>
            <Button variant="ghost" size="sm" @click="openCustom">Other</Button>
        </template>
        <template v-else>
            <Button variant="secondary" size="sm" @click="openCustom">
                <template #icon-left><Plus class="size-4" /></template>
                Log water
            </Button>
            <Button v-if="!isError" variant="ghost" size="sm" @click="router.push({ name: 'diet-water-presets' })">
                Add drink sizes
            </Button>
        </template>
        <IconButton class="ml-auto" size="sm" :label="`Water entries for ${dayLabel}`" @click="openEntries">
            <ListIcon />
        </IconButton>
    </div>
</template>
