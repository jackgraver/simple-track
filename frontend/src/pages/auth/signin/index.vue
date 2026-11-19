<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuth } from "~/composables/auth/useAuth";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { Button, Field, Input } from "~/shared/ui";

const router = useRouter();
const route = useRoute();
const { login, register } = useAuth();

function redirectAfterSignIn() {
    const r = route.query.redirect;
    if (typeof r === "string" && r.startsWith("/") && !r.startsWith("//")) {
        return r;
    }
    return "/";
}

const isLoginMode = ref(true);
// Vite dev server only: prefill the local dev account (created by `dev`, see cmd/devuser).
const username = ref(import.meta.env.DEV ? "dev" : "");
const password = ref(import.meta.env.DEV ? "dev" : "");
const email = ref("");
const isLoading = ref(false);
const error = ref("");

const handleSubmit = async () => {
    error.value = "";
    isLoading.value = true;

    try {
        if (isLoginMode.value) {
            await login(username.value, password.value);
            toast.push("Login successful!", "success");
            await router.push(redirectAfterSignIn());
        } else {
            await register(
                username.value,
                password.value,
                email.value || undefined,
            );
            toast.push("Registration successful!", "success");
            await router.push(redirectAfterSignIn());
        }
    } catch (err) {
        error.value = getApiErrorMessage(err, "An error occurred");
        toast.push(error.value, "error");
    } finally {
        isLoading.value = false;
    }
};
</script>

<template>
    <div class="flex min-h-dvh items-center justify-center bg-canvas p-4">
        <div class="w-full max-w-sm rounded-card border border-line bg-surface p-6">
            <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
                <p v-if="error" role="alert" class="m-0 text-sm text-danger">{{ error }}</p>

                <Field label="Username" required>
                    <Input v-model="username" required autocomplete="username" />
                </Field>

                <Field label="Password" required>
                    <Input v-model="password" type="password" required autocomplete="current-password" />
                </Field>

                <Field v-if="!isLoginMode" label="Email (optional)">
                    <Input v-model="email" type="email" inputmode="email" autocomplete="email" />
                </Field>

                <Button type="submit" variant="primary" size="lg" block :loading="isLoading">
                    {{ isLoginMode ? "Sign in" : "Sign up" }}
                </Button>
            </form>
        </div>
    </div>
</template>
