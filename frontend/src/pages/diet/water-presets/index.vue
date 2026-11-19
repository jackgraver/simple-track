<script setup lang="ts">
import axios from "axios";
import { computed, ref, watch } from "vue";
import type { DrinkSizePreset } from "~/api/tracking/types";
import {
    useCreateDrinkSizePreset,
    useDeleteDrinkSizePreset,
    useDrinkSizePresets,
    useUpdateDrinkSizePreset,
} from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { Pencil, Trash2 } from "lucide-vue-next";
import {
    Button,
    Card,
    EmptyState,
    ErrorState,
    Field,
    IconButton,
    Input,
    List,
    ListRow,
    PageHeader,
    Section,
    Select,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";

const { goalOz, displayUnit, ozFromDisplayAmount, formatVolumeFromOz } =
    useWaterPrefs();

const goalInput = ref("");

const newName = ref("");
const newAmount = ref("");

const editingId = ref<number | null>(null);
const editName = ref("");
const editAmount = ref("");

const {
    data: presets,
    isPending,
    isError,
    error,
    refetch,
} = useDrinkSizePresets();
const createMut = useCreateDrinkSizePreset();
const updateMut = useUpdateDrinkSizePreset();
const deleteMut = useDeleteDrinkSizePreset();

const rows = computed(() => presets.value ?? []);
const showSkeleton = useDelayedFlag(isPending);
const unitOptions = [
    { label: "oz", value: "oz" as const },
    { label: "ml", value: "ml" as const },
];

function syncGoalInputFromPrefs() {
    const f = formatVolumeFromOz(goalOz.value);
    goalInput.value = String(f.value);
}

syncGoalInputFromPrefs();

watch(displayUnit, () => {
    const f = formatVolumeFromOz(goalOz.value);
    goalInput.value = String(f.value);
});

function applyGoalFromInput() {
    const n = Number.parseFloat(goalInput.value.replace(/\s/g, ""));
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a positive goal", "error");
        syncGoalInputFromPrefs();
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    goalOz.value = oz;
    syncGoalInputFromPrefs();
    toast.push("Goal saved", "success");
}

async function createPreset() {
    const name = newName.value.trim();
    const n = Number.parseFloat(newAmount.value.replace(/\s/g, ""));
    if (!name) {
        toast.push("Name is required", "error");
        return;
    }
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    try {
        await createMut.mutateAsync({ name, amountOz: oz });
        toast.push("Size created", "success");
        newName.value = "";
        newAmount.value = "";
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

function startEdit(p: DrinkSizePreset) {
    editingId.value = p.ID;
    editName.value = p.name;
    const f = formatVolumeFromOz(p.amount_oz);
    editAmount.value = String(f.value);
}

function cancelEdit() {
    editingId.value = null;
}

async function saveEdit() {
    const id = editingId.value;
    if (id == null) return;
    const name = editName.value.trim();
    const n = Number.parseFloat(editAmount.value.replace(/\s/g, ""));
    if (!name) {
        toast.push("Name is required", "error");
        return;
    }
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    try {
        await updateMut.mutateAsync({ id, name, amountOz: oz });
        toast.push("Updated", "success");
        editingId.value = null;
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

async function removePreset(id: number) {
    const ok = await dialogManager.confirm({
        title: "Delete drink size?",
        tone: "danger",
        message: "Quick-add buttons will no longer include this size.",
        confirmText: "Delete",
        cancelText: "Cancel",
    });
    if (!ok) return;
    try {
        await deleteMut.mutateAsync(id);
        toast.push("Deleted", "success");
        if (editingId.value === id) editingId.value = null;
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

function axiosMsg(err: unknown): string {
    if (
        axios.isAxiosError(err) &&
        err.response?.data &&
        typeof err.response.data === "object" &&
        "error" in err.response.data
    ) {
        const e0 = (err.response.data as { error?: string }).error;
        if (e0) return e0;
    }
    if (err instanceof Error) return err.message;
    return "Request failed";
}

function displayAmountLabel(amountOz: number): string {
    const f = formatVolumeFromOz(amountOz);
    return `${f.value} ${f.unit}`;
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-3xl flex-col gap-6 pt-2 pb-10">
        <PageHeader title="Water sizes & goal">
            <template #actions>
                <Button variant="ghost" size="sm" :to="{ name: 'diet' }"
                    >Back to diet</Button
                >
            </template>
        </PageHeader>

        <Section title="Display & daily goal">
            <Card>
                <div class="flex flex-col gap-4">
                    <Field label="Unit">
                        <Select v-model="displayUnit" :options="unitOptions" />
                    </Field>
                    <Field
                        label="Daily goal"
                        :hint="`In ${displayUnit}. Amounts are stored in oz; switching units only changes how numbers are shown.`"
                    >
                        <Input
                            v-model="goalInput"
                            inputmode="decimal"
                            @change="applyGoalFromInput"
                            @keyup.enter.prevent="applyGoalFromInput"
                        >
                            <template #suffix>{{ displayUnit }}</template>
                        </Input>
                    </Field>
                    <div>
                        <Button variant="secondary" @click="applyGoalFromInput"
                            >Save goal</Button
                        >
                    </div>
                </div>
            </Card>
        </Section>

        <Section title="New drink size">
            <Card>
                <div class="flex flex-col gap-4">
                    <Field label="Label">
                        <Input
                            v-model="newName"
                            autocomplete="off"
                            placeholder="e.g. Large bottle"
                        />
                    </Field>
                    <Field label="Amount">
                        <Input
                            v-model="newAmount"
                            inputmode="decimal"
                            placeholder="e.g. 24"
                            @keyup.enter.prevent="createPreset"
                        >
                            <template #suffix>{{ displayUnit }}</template>
                        </Input>
                    </Field>
                    <div>
                        <Button
                            variant="primary"
                            :loading="createMut.isPending.value"
                            @click="createPreset"
                            >Add size</Button
                        >
                    </div>
                </div>
            </Card>
        </Section>

        <Section title="Saved sizes">
            <Card padding="none">
                <ErrorState
                    v-if="isError"
                    :error="error"
                    title="Couldn't load drink sizes"
                    @retry="refetch()"
                />
                <SkeletonList
                    v-else-if="showSkeleton"
                    :rows="3"
                    :leading="false"
                />
                <EmptyState
                    v-else-if="!isPending && !rows.length"
                    title="No sizes yet"
                    description="Add one above and it appears in the quick-add buttons."
                />
                <List v-else-if="rows.length">
                    <template v-for="p in rows" :key="p.ID">
                        <li
                            v-if="editingId === p.ID"
                            class="flex flex-col gap-3 px-3 py-3 lg:px-4"
                        >
                            <Field label="Label"
                                ><Input v-model="editName"
                            /></Field>
                            <Field label="Amount">
                                <Input v-model="editAmount" inputmode="decimal"
                                    ><template #suffix>{{
                                        displayUnit
                                    }}</template></Input
                                >
                            </Field>
                            <div class="flex gap-2">
                                <Button
                                    variant="primary"
                                    size="sm"
                                    :loading="updateMut.isPending.value"
                                    @click="saveEdit"
                                    >Save</Button
                                >
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    @click="cancelEdit"
                                    >Cancel</Button
                                >
                            </div>
                        </li>
                        <ListRow
                            v-else
                            :pending="
                                deleteMut.isPending.value &&
                                deleteMut.variables.value === p.ID
                            "
                        >
                            {{ p.name }}
                            <template #trailing>
                                <span class="mr-1">{{
                                    displayAmountLabel(p.amount_oz)
                                }}</span>
                                <IconButton
                                    label="Edit size"
                                    size="sm"
                                    @click="startEdit(p)"
                                    ><Pencil
                                /></IconButton>
                                <IconButton
                                    label="Delete size"
                                    size="sm"
                                    variant="danger"
                                    @click="removePreset(p.ID)"
                                    ><Trash2
                                /></IconButton>
                            </template>
                        </ListRow>
                    </template>
                </List>
            </Card>
        </Section>
    </div>
</template>
