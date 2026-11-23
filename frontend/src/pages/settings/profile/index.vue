<script setup lang="ts">
import axios from 'axios';
import { computed, ref, watch } from 'vue';
import { Button, Card, ErrorState, PageHeader, Section, Skeleton, useDelayedFlag } from '~/shared/ui';
import { useProfile, useSaveProfile } from '~/api/tracking/queries';
import { toast } from '~/composables/toast/useToast';
import AccountSection from '../AccountSection.vue';
import AppearanceSection from '../AppearanceSection.vue';
import BodyProfileForm from './BodyProfileForm.vue';

const { data: profileRow, isPending, isError, error, refetch } = useProfile();
const showSkeleton = useDelayedFlag(isPending);
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
            // Keep the level chosen on /diet/targets; this form doesn't edit it.
            activity_level: profileRow.value?.activity_level,
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
    <div class="mx-auto flex w-full max-w-md flex-col gap-6 pb-8">
        <PageHeader title="Settings" />
        <AppearanceSection />
        <Section title="Body profile">
            <p class="m-0 text-sm text-fg-muted">
                Used for TDEE estimates on macro targets. Weight comes from your latest weight log.
            </p>
            <ErrorState v-if="isError" :error="error" title="Couldn't load your profile" @retry="refetch()" />
            <div v-else-if="isPending" class="flex flex-col gap-3">
                <template v-if="showSkeleton">
                    <Skeleton shape="block" height="2.75rem" />
                    <Skeleton shape="block" height="2.75rem" />
                    <Skeleton shape="block" height="2.75rem" />
                </template>
            </div>
            <form v-else class="flex flex-col gap-4" @submit.prevent="save">
                <Card>
                    <BodyProfileForm
                        :height-in="heightIn"
                        :age="age"
                        :sex="sex"
                        @update:height-in="heightIn = $event"
                        @update:age="age = $event"
                        @update:sex="sex = $event"
                    />
                </Card>
                <Button type="submit" variant="primary" size="lg" block :loading="saving" :disabled="!profileValid">
                    Save profile
                </Button>
            </form>
        </Section>
        <AccountSection />
    </div>
</template>
