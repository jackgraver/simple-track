<script setup lang="ts">
import { onMounted } from "vue";
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import { useAuth } from "./composables/auth/useAuth";

const { getUsername } = useAuth();

onMounted(() => {
    document.documentElement.classList.add("dark-mode");
});
</script>

<template>
    <main
        class="dark-mode mx-auto box-border w-full max-w-full min-h-dvh overflow-x-clip"
    >
        <div
            v-if="getUsername()"
            class="flex items-center justify-end lg:pr-16 pb-2"
        >
            <div class="flex items-center gap-2 pt-2">
                <router-link :to="{ name: 'gym' }">
                    <p class="bg-gray-500 rounded-md p-2">Gym</p>
                </router-link>
                <router-link :to="{ name: 'diet' }">
                    <p class="bg-gray-500 rounded-md p-2">Diet</p>
                </router-link>
                <p>{{ getUsername() }}</p>
            </div>
        </div>
        <div
            class="flex w-full flex-col items-stretch gap-4 px-4 lg:px-0 lg:mx-auto lg:max-w-8/12"
        >
            <RouterView />
        </div>
    </main>
    <ToastContainer />
    <DialogContainer />
</template>

<style>
html {
    color: rgb(218, 218, 218);
    background-color: rgb(20, 20, 20);
    font-family:
        system-ui,
        -apple-system,
        BlinkMacSystemFont,
        "Segoe UI",
        Roboto,
        "Helvetica Neue",
        Arial,
        sans-serif;
    margin: 0;
}

body {
    margin: 0;
}

.grid-container {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-rows: auto 1fr;
    height: 100%;
    width: 100%;
    margin: 0;
}

.grid-top-left {
    grid-column: 1;
    grid-row: 1;
    border-right: 0.5px solid #a7a7a7;
    border-bottom: 0.5px solid #a7a7a7;
}

.grid-top-right {
    grid-column: 2;
    grid-row: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 1rem;
    border-bottom: 0.5px solid #a7a7a7;
}

.grid-top-right h1 {
    font-size: 1.5rem;
    margin: 0;
}

.current-date-label {
    font-size: 1rem;
}

.grid-bottom-left {
    grid-column: 1;
    grid-row: 2;
    overflow-y: auto;
    border-right: 0.5px solid #a7a7a7;
}

.grid-bottom-right {
    grid-column: 2;
    grid-row: 2;
    overflow-y: auto;
    padding: 1rem;
}

@media (max-width: 767px) {
    .grid-container {
        grid-template-columns: 1fr;
        grid-template-rows: auto auto 1fr;
    }

    .grid-top-left {
        display: none;
    }

    .grid-top-right {
        grid-column: 1;
        grid-row: 1;
        justify-content: center;
    }

    .grid-bottom-left {
        grid-column: 1;
        grid-row: 2;
    }

    .grid-bottom-right {
        grid-column: 1;
        grid-row: 3;
    }
}

a,
.router-link {
    all: unset;
    cursor: pointer;
}

a,
.router-link {
    text-decoration: none;
    color: inherit;
    font: inherit;
}

button {
    /* background: #2c2c2c; */
    padding: 6px 12px;
    margin: 4px 0px;
    border-radius: 4px;
    border: 2px solid #a7a7a7;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    /* color: white; */
    border: none;
    cursor: pointer;
}

button:disabled {
    background: #ccc;
}

button:disabled:hover {
    cursor: not-allowed;
}

/* button:hover:not(:disabled) {
    background: #525252;
} */

button svg {
    vertical-align: middle;
    pointer-events: none;
}

.delete-button {
    background: #ff2b2b;
}

.delete-button:hover:not(:disabled) {
    background: #c91e1e;
}

.confirm-button {
    background: #55db2c;
}

.confirm-button:hover:not(:disabled) {
    background: #38a716;
}

input,
select {
    background: rgb(71, 71, 71);
    color: white;
    border: none;
    padding: 6px 12px;
    margin: 4px 0 4px 0;
    border-radius: 4px;
}

input:focus {
    outline: none;
}
</style>
