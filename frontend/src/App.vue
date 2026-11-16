<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import NavBar from "./shared/NavBar.vue";
import DateAndBreadcrumbs from "./shared/DateAndBreadcrumbs.vue";

const route = useRoute();
const showDateAndBreadcrumbs = computed(
    () => route.meta.hideDateAndBreadcrumbs !== true,
);
</script>

<template>
    <main
        class="dark scrollbar-gutter-both mx-auto box-border w-full max-w-full min-h-dvh overflow-x-clip"
    >
        <div
            class="flex w-full flex-col items-stretch px-4 lg:px-8 xl:px-0 xl:mx-auto xl:max-w-8/12"
        >
            <NavBar />
            <DateAndBreadcrumbs v-if="showDateAndBreadcrumbs" />
            <RouterView />
        </div>
    </main>
    <ToastContainer />
    <DialogContainer />
</template>

<style>
html {
    color: rgb(218, 218, 218);
    background-color: var(--color-main-bg);
    margin: 0;
    min-height: 100%;
    overscroll-behavior: none;
}

body {
    margin: 0;
    padding-top: env(safe-area-inset-top);
    padding-bottom: env(safe-area-inset-bottom);
    overscroll-behavior: none;
}

@layer base {
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
}

button {
    padding: 6px 12px;
    margin: 4px 0px;
    border-radius: 4px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    border: none;
    cursor: pointer;
}

button:disabled:hover {
    cursor: not-allowed;
}

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

/* Keyboard focus indicator for links/buttons (the base reset above strips it). */
:focus-visible {
    outline: 2px solid var(--color-text-secondary);
    outline-offset: 2px;
}
</style>
