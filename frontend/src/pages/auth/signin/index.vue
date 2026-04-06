<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuth } from "~/composables/auth/useAuth";
import { toast } from "~/composables/toast/useToast";
import Input from "~/shared/input/Input.vue";

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
const username = ref("");
const password = ref("");
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
            // await register(
            //     username.value,
            //     password.value,
            //     email.value || undefined,
            // );
            // toast.push("Registration successful!", "success");
            // await router.push(redirectAfterSignIn());
        }
    } catch (err: any) {
        error.value = err.message || "An error occurred";
        toast.push(error.value, "error");
    } finally {
        isLoading.value = false;
    }
};

const toggleMode = () => {
    isLoginMode.value = !isLoginMode.value;
    error.value = "";
    password.value = "";
    email.value = "";
};
</script>

<template>
    <div class="signin-container">
        <div class="signin-card">
            <!-- <h1>{{ isLoginMode ? "Sign In" : "Sign Up" }}</h1> -->

            <form @submit.prevent="handleSubmit" class="signin-form">
                <div v-if="error" class="error-message">{{ error }}</div>

                <div class="form-group">
                    <Input
                        label=""
                        v-model="username"
                        type="text"
                        :required="true"
                        autocomplete="username"
                        placeholder=""
                        :error="error && !username ? error : ''"
                    />
                </div>

                <div class="form-group">
                    <Input
                        label=""
                        v-model="password"
                        type="password"
                        :required="true"
                        autocomplete="current-password"
                        placeholder=""
                        :error="error && !password ? error : ''"
                    />
                </div>

                <div class="form-group">
                    <Input
                        label=""
                        v-model="email"
                        type="text"
                        autocomplete=""
                        placeholder=""
                        :error="''"
                    />
                </div>

                <!-- <div class="form-group">
                    <Input
                        label="Username"
                        v-model="username"
                        type="text"
                        :required="true"
                        autocomplete="username"
                        placeholder="Enter your username"
                        :error="error && !username ? error : ''"
                    />
                </div>

                <div class="form-group">
                    <Input
                        label="Password"
                        v-model="password"
                        type="password"
                        :required="true"
                        autocomplete="current-password"
                        placeholder="Enter your password"
                        :error="error && !password ? error : ''"
                    />
                </div> -->

                <div v-if="!isLoginMode" class="form-group">
                    <Input
                        label="Email (optional)"
                        v-model="email"
                        type="email"
                        :required="false"
                        autocomplete="email"
                        placeholder="Enter your email"
                        :error="error && !email && !isLoginMode ? error : ''"
                    />
                </div>

                <button
                    type="submit"
                    :disabled="isLoading"
                    class="submit-button"
                >
                    {{
                        isLoading ? "Loading..." : isLoginMode ? "" : "Sign Up"
                    }}
                </button>
            </form>

            <!-- <div class="toggle-mode">
                <span>{{
                    isLoginMode
                        ? "Don't have an account?"
                        : "Already have an account?"
                }}</span>
                <button type="button" @click="toggleMode" class="toggle-button">
                    {{ isLoginMode ? "Sign Up" : "Sign In" }}
                </button>
            </div> -->
        </div>
    </div>
</template>

<style scoped>
.signin-container {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    padding: 1rem;
    background: rgb(20, 20, 20);
}

.signin-card {
    width: 100%;
    max-width: 400px;
    padding: 2rem;
    background: rgb(27, 27, 27);
    border: 1px solid rgb(56, 56, 56);
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
}

.signin-card h1 {
    margin: 0 0 2rem 0;
    text-align: center;
    color: white;
    font-size: 2rem;
}

.signin-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.submit-button {
    padding: 0.75rem 1.5rem;
    background: rgb(40, 80, 40);
    border: 1px solid rgb(56, 56, 56);
    border-radius: 4px;
    color: white;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
    margin-top: 0.5rem;
}

.submit-button:hover:not(:disabled) {
    background: rgb(50, 100, 50);
}

.submit-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.toggle-mode {
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid rgb(56, 56, 56);
    text-align: center;
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
}

.toggle-button {
    margin-left: 0.5rem;
    background: none;
    border: none;
    color: rgb(100, 150, 255);
    cursor: pointer;
    text-decoration: underline;
    font-size: 0.9rem;
    padding: 0;
}

.toggle-button:hover {
    color: rgb(150, 200, 255);
}
</style>
