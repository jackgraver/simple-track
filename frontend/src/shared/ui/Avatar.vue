<script setup lang="ts">
import { AvatarFallback, AvatarImage, AvatarRoot } from "reka-ui";

withDefaults(
    defineProps<{
        name: string;
        src?: string | null;
        size?: "sm" | "md" | "lg";
        /** Set false when the avatar stands alone, without the name beside it. */
        decorative?: boolean;
    }>(),
    { size: "md", decorative: true },
);

const sizes = { sm: "size-7", md: "size-9", lg: "size-16" } as const;
</script>

<template>
    <AvatarRoot
        :role="decorative ? undefined : 'img'"
        :aria-label="decorative ? undefined : name"
        :aria-hidden="decorative || undefined"
        class="relative inline-flex shrink-0 overflow-hidden rounded-full align-middle select-none"
        :class="sizes[size]"
    >
        <AvatarImage v-if="src" :src="src" alt="" class="size-full object-cover" />
        <AvatarFallback
            class="size-full bg-linear-to-br from-avatar-a to-avatar-b ring-1 ring-on-solid/25 ring-inset"
        >
            <span class="block size-full rounded-full bg-radial-[at_30%_25%] from-on-solid/45 to-transparent to-60%" />
        </AvatarFallback>
    </AvatarRoot>
</template>
