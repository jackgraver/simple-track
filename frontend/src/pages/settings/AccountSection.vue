<script setup lang="ts">
import { ref } from "vue";
import { Button, Section } from "~/shared/ui";
import { useAuth } from "~/composables/auth/useAuth";

const { username, logout } = useAuth();
const signingOut = ref(false);

// logout() never throws: a failed request still clears the session locally
// and goes to sign-in.
async function signOut() {
    signingOut.value = true;
    try {
        await logout();
    } finally {
        signingOut.value = false;
    }
}
</script>

<template>
    <Section title="Account">
        <p v-if="username" class="m-0 text-sm text-fg-muted">
            Signed in as <span class="font-medium text-fg">{{ username }}</span>
        </p>
        <Button variant="secondary" :loading="signingOut" @click="signOut">Sign out</Button>
    </Section>
</template>
