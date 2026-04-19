<script setup lang="ts">
import { useAuth } from "~/composables/auth/useAuth";

const { getUsername } = useAuth();
</script>

<template>
    <nav class="flex items-center justify-between pb-2 pt-1 w-full">
        <div class="pl-2 lg:pl-0 font-semibold text-lg tracking-tight">
            simpletrack
        </div>
        <div class="flex items-center gap-2 pt-2">
            <router-link :to="{ name: 'gym' }">
                <span class="hover:bg-secondBg rounded-md p-2 text-sm"
                    >Gym</span
                >
            </router-link>
            <router-link :to="{ name: 'diet' }">
                <span class="hover:bg-secondBg rounded-md p-2 text-sm"
                    >Diet</span
                >
            </router-link>
        </div>
        <div
            class="pr-2 lg:pr-0 min-w-[90px] text-right flex items-center justify-end"
        >
            <span v-if="getUsername()">{{ getUsername() }}</span>
        </div>
    </nav>
</template>
