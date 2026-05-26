<script setup lang="ts">
import { useAuth } from "~/composables/auth/useAuth";
import Breadcrumbs from "~/shared/Breadcrumbs.vue";
import TrackingNotifications from "~/shared/TrackingNotifications.vue";

const { getUsername } = useAuth();
</script>

<template>
    <nav class="flex items-center justify-between pb-2 pt-1 w-full">
        <Breadcrumbs class="min-w-0" />
        <div class="flex items-center gap-2 pt-2">
            <router-link
                :to="{
                    name: 'gym',
                    query: $route.query.offset
                        ? { offset: $route.query.offset }
                        : undefined,
                }"
                class="hover:bg-secondBg rounded-md p-2 text-sm"
                active-class="underline underline-offset-4"
                >Gym</router-link
            >
            <router-link
                :to="{
                    name: 'diet',
                    query: $route.query.offset
                        ? { offset: $route.query.offset }
                        : undefined,
                }"
                class="hover:bg-secondBg rounded-md p-2 text-sm"
                active-class="underline underline-offset-4"
                >Diet</router-link
            >
        </div>
        <div
            class="pr-2 lg:pr-0 min-w-[90px] flex items-center justify-end gap-2 text-right"
        >
            <TrackingNotifications />
            <router-link
                v-if="getUsername()"
                :to="{ name: 'settings-profile' }"
                class="text-sm hover:underline underline-offset-4"
            >{{ getUsername() }}</router-link>
        </div>
    </nav>
</template>
