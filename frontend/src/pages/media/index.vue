<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
    getMediaStats,
    listMedia,
    MEDIA_STATUS_LABELS,
    MEDIA_STATUSES,
    MEDIA_TYPE_LABELS,
    MEDIA_TYPES,
    type MediaItem,
    type MediaListParams,
    type MediaStats,
    type MediaStatus,
    type MediaType,
} from "~/api/media/api";
import { getApiErrorMessage } from "~/api/httpError";
import StarRating from "./StarRating.vue";

const items = ref<MediaItem[]>([]);
const stats = ref<MediaStats | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");

const status = ref<MediaStatus | "">("");
const type = ref<MediaType | "">("");
const search = ref("");
const sort = ref<NonNullable<MediaListParams["sort"]>>("updated");
const favoritesOnly = ref(false);

const statusTabs = computed(() => [
    { value: "" as const, label: "All", count: stats.value?.total ?? 0 },
    ...MEDIA_STATUSES.map((s) => ({
        value: s,
        label: MEDIA_STATUS_LABELS[s],
        count: stats.value?.by_status[s] ?? 0,
    })),
]);

let searchTimer: ReturnType<typeof setTimeout> | undefined;
let requestId = 0;

async function load() {
    const id = ++requestId;
    isLoading.value = true;
    errorMessage.value = "";
    try {
        const [list, s] = await Promise.all([
            listMedia({
                status: status.value || undefined,
                type: type.value || undefined,
                q: search.value.trim() || undefined,
                favorite: favoritesOnly.value || undefined,
                sort: sort.value,
            }),
            getMediaStats(),
        ]);
        if (id !== requestId) return;
        items.value = list;
        stats.value = s;
    } catch (error) {
        if (id !== requestId) return;
        errorMessage.value = getApiErrorMessage(error, "Failed to load media");
    } finally {
        if (id === requestId) isLoading.value = false;
    }
}

watch([status, type, sort, favoritesOnly], () => void load());
watch(search, () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => void load(), 250);
});

function progressLabel(item: MediaItem) {
    if (item.progress == null) return "";
    return item.length ? `${item.progress}/${item.length}` : `${item.progress}`;
}

onMounted(() => void load());
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <header class="flex flex-wrap items-start justify-between gap-4">
            <div>
                <h1 class="m-0 text-xl font-semibold text-textPrimary">Media</h1>
                <p class="mb-0 mt-1 text-sm text-textSecondary">
                    <template v-if="stats">
                        {{ stats.completed_this_year }} finished this year<template v-if="stats.average_rating != null">
                            · avg ★ {{ stats.average_rating.toFixed(1) }}</template>
                    </template>
                    <template v-else>Things you've watched, read and played</template>
                </p>
            </div>
            <router-link
                :to="{ name: 'media-new', query: status === 'planned' ? { status: 'planned' } : undefined }"
                class="rounded-md bg-primary px-3 py-2 text-sm text-white"
                >Add</router-link
            >
        </header>

        <nav class="-mx-1 flex gap-1 overflow-x-auto pb-1">
            <button
                v-for="tab in statusTabs"
                :key="tab.value"
                type="button"
                class="shrink-0 rounded-md px-3 py-1.5 text-sm"
                :class="status === tab.value ? 'bg-thirdBg text-textPrimary' : 'text-textSecondary hover:bg-secondBg'"
                @click="status = tab.value"
            >
                {{ tab.label }} <span class="text-xs opacity-60">{{ tab.count }}</span>
            </button>
        </nav>

        <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto_auto]">
            <input
                v-model="search"
                type="search"
                placeholder="Search title or creator"
                class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none"
            />
            <select v-model="type" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none">
                <option value="">All types</option>
                <option v-for="t in MEDIA_TYPES" :key="t" :value="t">{{ MEDIA_TYPE_LABELS[t] }}</option>
            </select>
            <select v-model="sort" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none">
                <option value="updated">Recently updated</option>
                <option value="created">Recently added</option>
                <option value="finished">Recently finished</option>
                <option value="rating">Rating</option>
                <option value="title">Title</option>
                <option value="release">Release year</option>
            </select>
            <label class="flex items-center gap-2 rounded-md bg-thirdBg px-3 py-2 text-sm">
                <input v-model="favoritesOnly" type="checkbox" /> Favourites
            </label>
        </div>

        <p v-if="errorMessage" class="m-0 text-sm text-(--color-cf-red)">{{ errorMessage }}</p>
        <p v-else-if="isLoading && !items.length" class="m-0 text-sm text-textSecondary">Loading…</p>
        <div
            v-else-if="!items.length"
            class="rounded-lg bg-secondBg p-4 text-sm text-textSecondary"
        >
            Nothing here yet.
            <router-link :to="{ name: 'media-new' }" class="text-textPrimary underline">Add something</router-link>.
        </div>
        <section v-else class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            <router-link
                v-for="item in items"
                :key="item.ID"
                :to="{ name: 'media-detail', params: { id: item.ID } }"
                class="group flex flex-col gap-2"
            >
                <div class="relative aspect-[2/3] overflow-hidden rounded-lg bg-secondBg">
                    <img
                        v-if="item.cover_url"
                        :src="item.cover_url"
                        :alt="item.title"
                        loading="lazy"
                        referrerpolicy="no-referrer"
                        class="h-full w-full object-cover transition group-hover:scale-[1.02]"
                    />
                    <div v-else class="flex h-full items-center justify-center p-3 text-center text-sm text-textSecondary">
                        {{ item.title }}
                    </div>
                    <span class="absolute left-1.5 top-1.5 rounded bg-black/60 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-white">
                        {{ MEDIA_TYPE_LABELS[item.type] }}
                    </span>
                    <span v-if="item.favorite" class="absolute right-1.5 top-1.5 text-rose-400" title="Favourite">♥</span>
                    <span
                        v-if="status === '' && item.status !== 'completed'"
                        class="absolute bottom-1.5 left-1.5 rounded bg-black/60 px-1.5 py-0.5 text-[10px] text-white"
                    >
                        {{ MEDIA_STATUS_LABELS[item.status] }}<template v-if="progressLabel(item)"> · {{ progressLabel(item) }}</template>
                    </span>
                </div>
                <div class="min-w-0">
                    <p class="m-0 truncate text-sm font-medium text-textPrimary">{{ item.title }}</p>
                    <p class="m-0 truncate text-xs text-textSecondary">
                        {{ [item.creator, item.release_year].filter(Boolean).join(" · ") || " " }}
                    </p>
                    <StarRating v-if="item.rating != null" :model-value="item.rating" readonly size="sm" />
                </div>
            </router-link>
        </section>
    </div>
</template>
