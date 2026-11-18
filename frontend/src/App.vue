<script setup lang="ts">
import { computed, watchEffect } from "vue";
import { useRoute } from "vue-router";
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import { isAuthenticated } from "./composables/auth/session";
import DateAndBreadcrumbs from "./shared/DateAndBreadcrumbs.vue";
import Breadcrumbs from "./shared/Breadcrumbs.vue";
import SideNav from "./shared/nav/SideNav.vue";
import MobileHeader from "./shared/nav/MobileHeader.vue";
import BottomTabBar from "./shared/nav/BottomTabBar.vue";
import { useNavMemory } from "./shared/nav/useNavMemory";

const route = useRoute();
useNavMemory();

const showNav = computed(
    () => isAuthenticated.value && route.meta.focus !== true,
);
const showDate = computed(() => route.meta.dated === true);

// Lets fixed elements (toasts) sit above the mobile tab bar.
watchEffect(() => {
    document.documentElement.classList.toggle("has-tab-bar", showNav.value);
});
</script>

<template>
    <div class="status-bar-cover" aria-hidden="true"></div>
    <div class="dark min-h-dvh w-full">
        <SideNav v-if="showNav" class="hidden! lg:flex!" />
        <main
            class="scrollbar-gutter-both box-border w-full overflow-x-clip"
            :class="{ 'pb-[4.5rem] lg:pb-0': showNav }"
        >
            <div
                class="mx-auto flex w-full max-w-6xl flex-col items-stretch px-4 lg:px-8"
                :class="{ 'app-content': showNav }"
            >
                <template v-if="showNav">
                    <MobileHeader class="lg:hidden" />
                    <Breadcrumbs class="hidden pt-4 pb-1 lg:flex" />
                </template>
                <DateAndBreadcrumbs
                    v-if="showDate"
                    :class="showNav ? 'pb-4 lg:pt-4' : 'py-4'"
                />
                <RouterView />
            </div>
        </main>
        <BottomTabBar v-if="showNav" class="lg:hidden" />
    </div>
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

/* Solid cover behind the iOS status bar / Dynamic Island so scrolled content
   (and iOS's own translucent edge effect) doesn't show through. */
.status-bar-cover {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: env(safe-area-inset-top);
    background-color: var(--color-main-bg);
    z-index: 40;
    pointer-events: none;
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
