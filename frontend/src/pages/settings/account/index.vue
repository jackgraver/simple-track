<script setup lang="ts">
import { computed, ref } from "vue";
import { changePassword, passwordProblem } from "~/api/account/api";
import { getApiErrorMessage } from "~/api/httpError";
import { username } from "~/composables/auth/session";
import { toast } from "~/composables/toast/useToast";
import { Button, Card, Field, Input, PageHeader, Section } from "~/shared/ui";

const current = ref("");
const next = ref("");
const confirm = ref("");
const submitted = ref(false);
const saving = ref(false);

const nextError = computed(() => {
    if (!submitted.value) return undefined;
    if (next.value && next.value === current.value) return "Choose a different password.";
    return passwordProblem(next.value) ?? undefined;
});
const confirmError = computed(() =>
    submitted.value && confirm.value !== next.value ? "The passwords don't match." : undefined,
);

async function save() {
    submitted.value = true;
    if (!current.value || nextError.value || confirmError.value) return;
    saving.value = true;
    try {
        await changePassword(current.value, next.value);
        current.value = next.value = confirm.value = "";
        submitted.value = false;
        toast.push("Password changed. Other devices are signed out.", "success");
    } catch (err) {
        toast.push(getApiErrorMessage(err, "Couldn't change the password"), "error");
    } finally {
        saving.value = false;
    }
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-md flex-col gap-6 pb-8">
        <PageHeader title="Account">
            <template #meta>Signed in as {{ username }}</template>
        </PageHeader>
        <Section title="Change password">
            <form class="flex flex-col gap-4" novalidate @submit.prevent="save">
                <Card>
                    <div class="flex flex-col gap-4">
                        <!-- eslint-disable-next-line vue/no-restricted-html-elements -- hidden username field for password managers, not a control -->
                        <input type="text" class="sr-only" autocomplete="username" :value="username ?? ''" readonly tabindex="-1" aria-hidden="true" />
                        <Field label="Current password" :error="submitted && !current ? 'Enter your current password.' : undefined" required>
                            <Input v-model="current" type="password" autocomplete="current-password" :invalid="submitted && !current" required />
                        </Field>
                        <Field label="New password" hint="At least 10 characters. Signs out your other devices." :error="nextError" required>
                            <Input v-model="next" type="password" autocomplete="new-password" :invalid="!!nextError" required />
                        </Field>
                        <Field label="Repeat new password" :error="confirmError" required>
                            <Input v-model="confirm" type="password" autocomplete="new-password" :invalid="!!confirmError" required />
                        </Field>
                    </div>
                </Card>
                <Button type="submit" variant="primary" size="lg" block :loading="saving">Change password</Button>
            </form>
        </Section>
    </div>
</template>
