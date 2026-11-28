<script setup lang="ts">
// /setup/:token: the one-time link an admin texts to set (new account) or reset a password.
// Public: no session needed. Setting the password signs this device in as that user.
import axios from "axios";
import { KeyRound, Link2Off } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuery } from "@tanstack/vue-query";
import { completeSetup, inspectSetupLink, passwordProblem } from "~/api/account/api";
import { getApiErrorMessage } from "~/api/httpError";
import { queryClient } from "~/api/queryClient";
import { applyAuthResponse } from "~/composables/auth/useAuth";
import { toast } from "~/composables/toast/useToast";
import { Button, Card, EmptyState, ErrorState, Field, Input, Skeleton, useDelayedFlag } from "~/shared/ui";

const route = useRoute();
const router = useRouter();
const token = computed(() => String(route.params.token ?? ""));

// Not cached: the link's state is what matters, and it changes once it's used.
const { data: info, isPending, isError, error, refetch } = useQuery({
    queryKey: computed(() => ["setup-link", token.value]),
    queryFn: () => inspectSetupLink(token.value),
    retry: false,
    gcTime: 0,
    staleTime: Infinity,
});
const showSkeleton = useDelayedFlag(isPending);
const linkDead = computed(() => axios.isAxiosError(error.value) && error.value.response?.status === 410);

const password = ref("");
const confirm = ref("");
const submitted = ref(false);
const saving = ref(false);
const deadOnSubmit = ref(false);

const passwordError = computed(() => (submitted.value ? passwordProblem(password.value) ?? undefined : undefined));
const confirmError = computed(() =>
    submitted.value && confirm.value !== password.value ? "The passwords don't match." : undefined,
);
const isReset = computed(() => info.value?.purpose === "reset");

async function submit() {
    submitted.value = true;
    if (passwordError.value || confirmError.value) return;
    saving.value = true;
    try {
        const data = await completeSetup(token.value, password.value);
        // Whoever used this device before is signed out now; drop their cached data.
        queryClient.clear();
        applyAuthResponse(data);
        toast.push(isReset.value ? "Password changed. You're signed in." : "Welcome! You're signed in.", "success");
        await router.replace("/");
    } catch (err) {
        if (axios.isAxiosError(err) && err.response?.status === 410) {
            deadOnSubmit.value = true;
        } else {
            toast.push(getApiErrorMessage(err, "Couldn't set the password"), "error");
        }
    } finally {
        saving.value = false;
    }
}
</script>

<template>
    <div class="flex min-h-dvh items-center justify-center bg-canvas p-4">
        <Card class="w-full max-w-sm">
            <div v-if="isPending" class="flex flex-col gap-4" aria-busy="true">
                <template v-if="showSkeleton">
                    <Skeleton shape="block" height="1.5rem" />
                    <Skeleton shape="block" height="2.75rem" />
                    <Skeleton shape="block" height="2.75rem" />
                </template>
            </div>
            <EmptyState
                v-else-if="linkDead || deadOnSubmit"
                :icon="Link2Off"
                title="This link no longer works"
                description="It has expired or was already used. Ask for a new one."
            >
                <template #action>
                    <Button variant="secondary" :to="{ name: 'signin' }">Go to sign in</Button>
                </template>
            </EmptyState>
            <ErrorState v-else-if="isError" :error="error" title="Couldn't check this link" @retry="refetch()" />
            <form v-else-if="info" class="flex flex-col gap-4" novalidate @submit.prevent="submit">
                <div class="flex flex-col gap-1">
                    <h1 class="m-0 flex items-center gap-2 text-lg font-semibold">
                        <KeyRound aria-hidden="true" class="size-5 text-fg-muted" />
                        {{ isReset ? "Reset your password" : "Set up your account" }}
                    </h1>
                    <p class="m-0 text-sm text-fg-muted">
                        Choose a password for <span class="font-medium text-fg">{{ info.username }}</span>.
                        You'll sign in with this username from now on.
                    </p>
                </div>
                <!-- Lets password managers save the username with the new password. -->
                <!-- eslint-disable-next-line vue/no-restricted-html-elements -- hidden username field for password managers, not a control -->
                <input type="text" class="sr-only" autocomplete="username" :value="info.username" readonly tabindex="-1" aria-hidden="true" />
                <Field label="New password" hint="At least 10 characters." :error="passwordError" required>
                    <Input v-model="password" type="password" autocomplete="new-password" :invalid="!!passwordError" required />
                </Field>
                <Field label="Repeat password" :error="confirmError" required>
                    <Input v-model="confirm" type="password" autocomplete="new-password" :invalid="!!confirmError" required />
                </Field>
                <Button type="submit" variant="primary" size="lg" block :loading="saving">
                    {{ isReset ? "Change password and sign in" : "Set password and sign in" }}
                </Button>
            </form>
        </Card>
    </div>
</template>
