<script setup lang="ts">
import { computed } from "vue";
import type { Food } from "~/types/diet";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { Badge } from "~/shared/ui";

const props = defineProps<{
    item: Food & { entry_kind?: string };
}>();

const isComposite = computed(() => props.item.entry_kind === "composite");
</script>

<template>
    <div class="flex min-w-0 flex-col items-start gap-1 whitespace-normal text-fg">
        <h3 class="m-0 flex items-center gap-2 text-base font-semibold">
            {{ item.name }}
            <Badge v-if="isComposite" tone="info">Recipe</Badge>
        </h3>
        <SimpleMacros
            class="flex-wrap gap-x-3"
            :calories="item.calories"
            :protein="item.protein"
            :fat="item.fat"
            :carbs="item.carbs"
            :fiber="item.fiber"
        />
    </div>
</template>
