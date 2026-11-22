<script setup lang="ts" generic="T extends string | number">
// Native <select>: the OS picker wins on mobile (PRIMITIVES). Same skin as Input.
import { computed } from "vue";
import { ChevronDown } from "lucide-vue-next";
import { useField } from "./fieldContext";

const model = defineModel<T | "">({ required: true });
const props = withDefaults(
    defineProps<{
        // `group` renders the option inside an <optgroup>; consecutive options with the same group share one.
        options: { label: string; value: T; group?: string }[];
        placeholder?: string;
        size?: "md" | "sm";
        invalid?: boolean;
        disabled?: boolean;
        required?: boolean;
        ariaLabel?: string;
    }>(),
    { size: "md" },
);
const field = useField();

const sections = computed(() => {
    const out: { group?: string; options: typeof props.options }[] = [];
    for (const o of props.options) {
        const last = out[out.length - 1];
        if (last && last.group === o.group) last.options.push(o);
        else out.push({ group: o.group, options: [o] });
    }
    return out;
});
</script>

<template>
    <div class="relative">
        <select
            :id="field?.id"
            v-model="model"
            :disabled="disabled"
            :required="required"
            :aria-label="ariaLabel"
            :aria-invalid="props.invalid || field?.invalid() || undefined"
            :aria-describedby="field?.describedBy()"
            class="w-full cursor-pointer appearance-none rounded-control border bg-surface-2 py-0 pr-9 pl-3 text-base text-fg outline-none transition-colors duration-150 focus:ring-1 disabled:cursor-not-allowed disabled:opacity-50 lg:text-sm"
            :class="[
                props.size === 'sm' ? 'min-h-touch lg:min-h-control-sm' : 'min-h-touch lg:min-h-row',
                props.invalid || field?.invalid()
                    ? 'border-danger ring-danger focus:border-danger'
                    : 'border-line-strong ring-accent focus:border-accent',
                model === '' ? 'text-fg-subtle' : '',
            ]"
        >
            <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
            <template v-for="(sec, si) in sections" :key="si">
                <optgroup v-if="sec.group" :label="sec.group">
                    <option v-for="o in sec.options" :key="String(o.value)" :value="o.value">{{ o.label }}</option>
                </optgroup>
                <template v-else>
                    <option v-for="o in sec.options" :key="String(o.value)" :value="o.value">{{ o.label }}</option>
                </template>
            </template>
        </select>
        <ChevronDown
            aria-hidden="true"
            class="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-fg-muted"
        />
    </div>
</template>
