<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import {
    MEDIA_STATUS_LABELS,
    MEDIA_STATUSES,
    MEDIA_TYPE_LABELS,
    MEDIA_TYPES,
    type MediaItem,
    type MediaListParams,
    type MediaStatus,
    type MediaType,
} from "~/api/media/api";
import { useMediaList, useMediaStats } from "~/api/media/queries";
import { Heart } from "lucide-vue-next";
import {
    Badge,
    Button,
    Checkbox,
    EmptyState,
    ErrorState,
    Input,
    PageHeader,
    Select,
    Skeleton,
    useDelayedFlag,
} from "~/shared/ui";
import StarRating from "./StarRating.vue";

const status = ref<MediaStatus | "">("");
const type = ref<MediaType | "">("");
const search = ref("");
const sort = ref<NonNullable<MediaListParams["sort"]>>("updated");
const favoritesOnly = ref(false);

const typeOptions = [
    { label: "All types", value: "" as const },
    ...MEDIA_TYPES.map((t) => ({ label: MEDIA_TYPE_LABELS[t], value: t })),
];
const sortOptions = [
    { label: "Recently updated", value: "updated" as const },
    { label: "Recently added", value: "created" as const },
    { label: "Recently finished", value: "finished" as const },
    { label: "Rating", value: "rating" as const },
    { label: "Title", value: "title" as const },
    { label: "Release year", value: "release" as const },
];
// Debounced copy of the search box drives the query key; filters change the key, cached lists show instantly.
const debouncedSearch = ref("");
const params = computed<MediaListParams>(() => ({
    status: status.value || undefined,
    type: type.value || undefined,
    q: debouncedSearch.value.trim() || undefined,
    favorite: favoritesOnly.value || undefined,
    sort: sort.value,
}));
const listQuery = useMediaList(params);
const statsQuery = useMediaStats();
const items = computed<MediaItem[]>(() => listQuery.data.value ?? []);
const stats = computed(() => statsQuery.data.value ?? null);
// Skeleton only on a true cold load (no data at all).
const isLoading = computed(() => listQuery.isPending.value);
const isRefreshing = computed(() => listQuery.isPlaceholderData.value);
const hasError = computed(() => listQuery.isError.value && listQuery.data.value === undefined);
const loadError = computed(() => listQuery.error.value);
const showSkeleton = useDelayedFlag(isLoading);

const statusTabs = computed(() => [
    { value: "" as const, label: "All", count: stats.value?.total ?? 0 },
    ...MEDIA_STATUSES.map((s) => ({
        value: s,
        label: MEDIA_STATUS_LABELS[s],
        count: stats.value?.by_status[s] ?? 0,
    })),
]);

let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(search, (value) => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => (debouncedSearch.value = value), 250);
});
onBeforeUnmount(() => clearTimeout(searchTimer));

function load() {
    void listQuery.refetch();
    void statsQuery.refetch();
}

function progressLabel(item: MediaItem) {
    if (item.progress == null) return "";
    return item.length ? `${item.progress}/${item.length}` : `${item.progress}`;
}

</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <PageHeader title="Media">
            <template #meta>
                <template v-if="stats">
                    {{ stats.completed_this_year }} finished this year<template v-if="stats.average_rating != null">
                        · avg ★ {{ stats.average_rating.toFixed(1) }}</template>
                </template>
                <template v-else>Things you've watched, read and played</template>
            </template>
            <template #actions>
                <Button
                    variant="primary"
                    :to="{ name: 'media-new', query: status === 'planned' ? { status: 'planned' } : undefined }"
                    >Add</Button
                >
            </template>
        </PageHeader>

        <nav aria-label="Status" class="-mx-4 flex gap-1 overflow-x-auto px-4 pb-1 lg:mx-0 lg:px-0">
            <Button
                v-for="tab in statusTabs"
                :key="tab.value"
                :variant="status === tab.value ? 'secondary' : 'ghost'"
                size="sm"
                :aria-pressed="status === tab.value"
                @click="status = tab.value"
            >
                {{ tab.label }} <span class="text-xs text-fg-subtle tabular-nums">{{ tab.count }}</span>
            </Button>
        </nav>

        <div class="grid gap-2 lg:grid-cols-[1fr_auto_auto_auto]">
            <Input v-model="search" type="search" placeholder="Search title or creator" aria-label="Search title or creator" />
            <Select v-model="type" :options="typeOptions" aria-label="Type" />
            <Select v-model="sort" :options="sortOptions" aria-label="Sort by" />
            <Checkbox v-model="favoritesOnly" label="Favourites" />
        </div>

        <ErrorState v-if="hasError" :error="loadError" title="Couldn't load media" @retry="load" />
        <div v-else-if="showSkeleton" class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            <div v-for="n in 10" :key="n" class="flex flex-col gap-2">
                <div class="aspect-[2/3] w-full"><Skeleton shape="block" width="100%" height="100%" /></div>
                <Skeleton width="75%" />
            </div>
        </div>
        <EmptyState
            v-else-if="!isLoading && !items.length"
            title="Nothing here yet"
            description="Add something you've watched, read or played."
        >
            <template #action>
                <Button variant="primary" :to="{ name: 'media-new' }">Add something</Button>
            </template>
        </EmptyState>
        <section v-else-if="items.length" class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5" :class="isRefreshing ? 'opacity-60' : ''">
            <router-link
                v-for="item in items"
                :key="item.ID"
                :to="{ name: 'media-detail', params: { id: item.ID } }"
                class="group flex flex-col gap-2 no-underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
                <div class="relative aspect-[2/3] overflow-hidden rounded-card border border-line bg-surface-2">
                    <img
                        v-if="item.cover_url"
                        :src="item.cover_url"
                        :alt="item.title"
                        loading="lazy"
                        referrerpolicy="no-referrer"
                        class="h-full w-full object-cover transition group-hover:scale-[1.02]"
                    />
                    <div v-else class="flex h-full items-center justify-center p-3 text-center text-sm text-fg-muted">
                        {{ item.title }}
                    </div>
                    <Badge class="absolute top-1.5 left-1.5 uppercase">{{ MEDIA_TYPE_LABELS[item.type] }}</Badge>
                    <span v-if="item.favorite" class="absolute top-1.5 right-1.5 text-danger">
                        <Heart aria-hidden="true" class="size-4 fill-current" />
                        <span class="sr-only">Favourite</span>
                    </span>
                    <Badge v-if="status === '' && item.status !== 'completed'" class="absolute bottom-1.5 left-1.5">
                        {{ MEDIA_STATUS_LABELS[item.status] }}<template v-if="progressLabel(item)"> · {{ progressLabel(item) }}</template>
                    </Badge>
                </div>
                <div class="min-w-0">
                    <p class="m-0 truncate text-sm font-medium text-fg">{{ item.title }}</p>
                    <p class="m-0 truncate text-xs text-fg-muted">
                        {{ [item.creator, item.release_year].filter(Boolean).join(" · ") || " " }}
                    </p>
                    <StarRating v-if="item.rating != null" :model-value="item.rating" readonly size="sm" />
                </div>
            </router-link>
        </section>
    </div>
</template>
