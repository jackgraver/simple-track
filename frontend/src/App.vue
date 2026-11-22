<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import { isAuthenticated } from "./composables/auth/session";
import DateAndBreadcrumbs from "./shared/DateAndBreadcrumbs.vue";
import OfflineNotice from "./shared/OfflineNotice.vue";
// Breadcrumbs (./shared/Breadcrumbs.vue) are kept but not displayed for now.
import SideNav from "./shared/nav/SideNav.vue";
import MobileHeader from "./shared/nav/mobile/MobileHeader.vue";
import BottomTabBar from "./shared/nav/mobile/BottomTabBar.vue";
import { useNavMemory } from "./shared/nav/mobile/useNavMemory";

const route = useRoute();
useNavMemory();

const showNav = computed(() => isAuthenticated.value);
// Focus pages (e.g. exercise logging) bring their own header, so only the
// mobile top bar is dropped; tab bar and sidebar stay for navigation.
const showMobileHeader = computed(
    () => showNav.value && route.meta.focus !== true,
);
const showDate = computed(() => route.meta.dated === true);
const wide = computed(() => route.meta.wide === true);

</script>

<template>
    <!-- Solid cover behind the iOS status bar / Dynamic Island so scrolled content doesn't show through. -->
    <div class="pointer-events-none fixed inset-x-0 top-0 z-nav h-safe-top bg-canvas" aria-hidden="true"></div>
    <div class="min-h-dvh w-full">
        <SideNav v-if="showNav" class="max-lg:hidden" />
        <main
            class="w-full overflow-x-clip"
            :class="{ 'pb-[4.5rem] lg:pb-0': showNav }"
        >
            <div
                class="flex w-full flex-col items-stretch"
                :class="[
                    showNav ? 'max-lg:mx-auto max-lg:max-w-6xl max-lg:px-4 lg:app-content' : 'mx-auto max-w-6xl px-4 lg:px-8',
                    showNav && wide ? 'lg:app-content-wide' : '',
                ]"
            >
                <MobileHeader v-if="showMobileHeader" class="lg:hidden" />
                <DateAndBreadcrumbs
                    v-if="showDate"
                    :class="showNav ? 'pb-4 lg:pt-4' : 'py-4'"
                />
                <OfflineNotice v-if="showNav" :class="{ 'mt-3': route.meta.focus === true }" />
                <RouterView />
            </div>
        </main>
        <BottomTabBar v-if="showNav" class="lg:hidden" />
    </div>
    <ToastContainer :tab-bar="showNav" :top="route.meta.focus === true" />
    <DialogContainer />
</template>
