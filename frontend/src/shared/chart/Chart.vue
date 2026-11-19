<script setup lang="ts">
// One chart.js wrapper for the whole app. Registers only what the app draws (no `chart.js/auto`).
// Colours come from the caller (via chartTheme()), so a theme switch only needs new options.
import {
    BarController,
    BarElement,
    CategoryScale,
    Chart as ChartJS,
    Filler,
    Legend,
    LinearScale,
    LineController,
    LineElement,
    PointElement,
    Tooltip,
    type ChartData,
    type ChartOptions,
} from "chart.js";
import { onBeforeUnmount, onMounted, ref, toRaw, watch } from "vue";

ChartJS.register(
    BarController, BarElement, LineController, LineElement, PointElement,
    CategoryScale, LinearScale, Filler, Legend, Tooltip,
);

type Kind = "line" | "bar";
const props = defineProps<{
    type: Kind;
    data: ChartData<Kind>;
    options: ChartOptions<Kind>;
    /** Accessible summary, e.g. "Body weight, last 30 days". */
    label?: string;
}>();

const canvas = ref<HTMLCanvasElement | null>(null);
let chart: ChartJS | null = null;

function create() {
    if (!canvas.value) return;
    chart = new ChartJS(canvas.value, {
        type: props.type,
        data: toRaw(props.data),
        options: toRaw(props.options),
    });
}

onMounted(create);
onBeforeUnmount(() => {
    chart?.destroy();
    chart = null;
});
watch(
    () => [props.data, props.options] as const,
    ([data, options]) => {
        if (!chart) return;
        chart.data = toRaw(data);
        chart.options = toRaw(options);
        chart.update();
    },
);
watch(
    () => props.type,
    () => {
        chart?.destroy();
        create();
    },
);
</script>

<template>
    <div class="relative">
        <canvas ref="canvas" role="img" :aria-label="label"></canvas>
    </div>
</template>
