<script setup lang="ts">
import axios from 'axios';
import { computed, ref, watch } from 'vue';
import { useProfile, useSaveProfile } from '~/api/tracking/queries';
import { toast } from '~/composables/toast/useToast';
import BodyProfileForm from './BodyProfileForm.vue';

const { data: profileRow, isPending, isError, error } = useProfile();
const saveMutation = useSaveProfile();
const heightIn = ref(70);
const age = ref(30);
const sex = ref<'male' | 'female'>('male');
watch(
    profileRow,
    (p) => {
        if (!p) return;
        heightIn.value = p.height_in;
        age.value = p.age;
        sex.value = p.sex;
    },
    { immediate: true },
);
const profileValid = computed(() => heightIn.value > 0 && age.value > 0);
const saving = computed(() => saveMutation.isPending.value);
async function save() {
    if (!profileValid.value) return;
    try {
        await saveMutation.mutateAsync({
            height_in: heightIn.value,
            age: age.value,
            sex: sex.value,
        });
        toast.push('Profile saved', 'success');
    } catch (err: unknown) {
        const msg =
            axios.isAxiosError(err) && typeof err.response?.data === 'object'
                ? String((err.response.data as { error?: string }).error ?? '')
                : '';
        toast.push(msg || 'Could not save profile', 'error');
    }
}
</script>

<template>
    <div class="flex flex-col gap-5 pb-8 pt-2">
        <h1 class="m-0 text-lg font-semibold text-textPrimary">Body profile</h1>
        <p class="m-0 text-sm text-zinc-500">
            Used for TDEE estimates on macro targets. Weight comes from your latest weight log.
        </p>
        <div v-if="isPending" class="text-zinc-500">Loading…</div>
        <div v-else-if="isError" class="text-red-400">
            {{ error?.message ?? 'Failed to load profile' }}
        </div>
        <form v-else class="flex max-w-md flex-col gap-4" @submit.prevent="save">
            <section class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4">
                <BodyProfileForm
                    :height-in="heightIn"
                    :age="age"
                    :sex="sex"
                    @update:height-in="heightIn = $event"
                    @update:age="age = $event"
                    @update:sex="sex = $event"
                />
            </section>
            <button
                type="submit"
                class="rounded-md border border-amber-700/50 bg-amber-950/40 px-4 py-2.5 text-sm font-semibold text-amber-100 transition-colors hover:bg-amber-950/70 disabled:opacity-50"
                :disabled="saving || !profileValid"
            >
                {{ saving ? 'Saving…' : 'Save profile' }}
            </button>
        </form>
    </div>
</template>
