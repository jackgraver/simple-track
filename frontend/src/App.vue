<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import { isAuthenticated } from "./composables/auth/session";
import DateAndBreadcrumbs from "./shared/DateAndBreadcrumbs.vue";
// Breadcrumbs (./shared/Breadcrumbs.vue) are kept but not displayed for now.
import SideNav from "./shared/nav/SideNav.vue";
import MobileHeader from "./shared/nav/mobile/MobileHeader.vue";
import MobileDrawer from "./shared/nav/mobile/MobileDrawer.vue";
import { useDrawerGesture } from "./shared/nav/mobile/useDrawerGesture";
import { MotionConfig, useIsDesktop } from "./shared/ui";

const route = useRoute();
useDrawerGesture(useIsDesktop());

const showNav = computed(() => isAuthenticated.value);
// Focus pages (e.g. exercise logging) bring their own header, so the mobile top bar
// (which holds the menu button for the nav drawer) is dropped there.
const showMobileHeader = computed(
    () => showNav.value && route.meta.focus !== true,
);
const showDate = computed(() => route.meta.dated === true);
const wide = computed(() => route.meta.wide === true);

</script>

<template>
    <!-- JS animations (motion-v) follow the OS reduced-motion setting, like the CSS ones. -->
    <MotionConfig reduced-motion="user">
        <!-- Solid cover behind the iOS status bar / Dynamic Island so scrolled content doesn't show through. -->
        <div class="pointer-events-none fixed inset-x-0 top-0 z-nav h-safe-top bg-canvas" aria-hidden="true"></div>
        <div class="min-h-dvh w-full">
            <SideNav v-if="showNav" class="max-lg:hidden" />
            <main class="w-full overflow-x-clip">
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
                    <RouterView />
                </div>
            </main>
        </div>
        <MobileDrawer v-if="showNav" />
        <ToastContainer :top="route.meta.focus === true" />
        <DialogContainer />
    </MotionConfig>
</template>
