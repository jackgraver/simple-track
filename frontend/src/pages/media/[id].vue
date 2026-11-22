<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
    LOOKUP_SOURCE_LABELS,
    LOOKUP_TYPES,
    lookupMedia,
    lookupMediaDetails,
    MEDIA_STATUS_LABELS,
    MEDIA_STATUSES,
    MEDIA_TYPE_LABELS,
    MEDIA_TYPES,
    type LookupResult,
    type MediaInput,
    type MediaItem,
    type MediaStatus,
    type MediaType,
} from "~/api/media/api";
import { useCreateMedia, useDeleteMedia, useMediaItem, useQuickUpdateMedia, useUpdateMedia } from "~/api/media/queries";
import { getApiErrorMessage } from "~/api/httpError";
import { useDialog } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import {
    Button,
    Card,
    Checkbox,
    EmptyState,
    ErrorState,
    Field,
    Input,
    List,
    ListRow,
    PageHeader,
    SegmentedControl,
    Select,
    Skeleton,
    Spinner,
    Textarea,
    useDelayedFlag,
} from "~/shared/ui";
import StarRating from "./StarRating.vue";

const route = useRoute();
const router = useRouter();
const { confirm } = useDialog();

const isNew = computed(() => route.name === "media-new");
const itemId = computed(() => Number(route.params.id));

// NaN id (new item) disables the query. The list cache seeds the detail, so most visits skip the cold load.
const itemQuery = useMediaItem(() => (isNew.value ? NaN : itemId.value));
const createMut = useCreateMedia();
const updateMut = useUpdateMedia();
const quickMut = useQuickUpdateMedia();
const deleteMut = useDeleteMedia();

const isLoading = computed(() => !isNew.value && itemQuery.data.value === undefined && !itemQuery.isError.value);
const hasError = computed(() => !isNew.value && itemQuery.isError.value && itemQuery.data.value === undefined);
const loadError = computed(() => itemQuery.error.value);
const isSaving = computed(() => createMut.isPending.value || updateMut.isPending.value);
const isDeleting = computed(() => deleteMut.isPending.value);
const showSkeleton = useDelayedFlag(isLoading);

const LENGTH_UNITS: Record<MediaType, string> = {
    movie: "minutes",
    tv: "episodes",
    book: "pages",
    game: "hours",
    anime: "episodes",
    comic: "issues",
    album: "tracks",
    podcast: "episodes",
    other: "units",
};

// v-model on <input type="number"> stores a number once edited, "" when cleared.
type NumberField = string | number;

const form = reactive({
    type: "movie" as MediaType,
    title: "",
    creator: "",
    release_year: "" as NumberField,
    genres: "",
    cover_url: "",
    synopsis: "",
    status: (route.query.status as MediaStatus) ?? "planned",
    rating: null as number | null,
    review: "",
    favorite: false,
    progress: "" as NumberField,
    length: "" as NumberField,
    started_at: "",
    finished_at: "",
    times_consumed: "0" as NumberField,
    source: "",
    external_id: "",
});

const typeOptions = MEDIA_TYPES.map((t) => ({ label: MEDIA_TYPE_LABELS[t], value: t }));
const statusOptions = MEDIA_STATUSES.map((s) => ({ label: MEDIA_STATUS_LABELS[s], value: s }));

const unit = computed(() => LENGTH_UNITS[form.type]);
const coverOk = ref(true);

// --- Online lookup: search the title box, pick a match, fill the form ---
const canLookup = computed(() => LOOKUP_TYPES.has(form.type));
const lookupResults = ref<LookupResult[] | null>(null);
const lookupError = ref("");
const isLookingUp = ref(false);
const applyingId = ref<string | null>(null);

function closeLookup() {
    lookupResults.value = null;
    lookupError.value = "";
}

async function runLookup() {
    const q = form.title.trim();
    if (!q || !canLookup.value || isLookingUp.value) return;
    isLookingUp.value = true;
    lookupError.value = "";
    try {
        lookupResults.value = await lookupMedia(form.type, q);
    } catch (error) {
        lookupResults.value = [];
        lookupError.value = getApiErrorMessage(error, "Lookup failed");
    } finally {
        isLookingUp.value = false;
    }
}

async function applyLookup(result: LookupResult) {
    if (applyingId.value) return;
    applyingId.value = result.external_id;
    try {
        const d = await lookupMediaDetails(result.source, result.external_id);
        // Overwrite descriptive fields only; status, rating, thoughts and
        // dates are yours. Nothing is saved until you press Save.
        Object.assign(form, {
            title: d.title || form.title,
            creator: d.creator || form.creator,
            release_year: d.release_year?.toString() ?? form.release_year,
            genres: d.genres || form.genres,
            cover_url: d.cover_url || form.cover_url,
            synopsis: d.synopsis || form.synopsis,
            length: d.length?.toString() ?? form.length,
            source: d.source,
            external_id: d.external_id,
        });
        coverOk.value = true;
        closeLookup();
        toast.push(
            `Filled in from ${LOOKUP_SOURCE_LABELS[d.source] ?? d.source}. Review and ${isNew.value ? "add" : "save"}.`,
            "success",
        );
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to load details"), "error");
    } finally {
        applyingId.value = null;
    }
}

function toForm(item: MediaItem) {
    Object.assign(form, {
        type: item.type,
        title: item.title,
        creator: item.creator,
        release_year: item.release_year?.toString() ?? "",
        genres: item.genres,
        cover_url: item.cover_url,
        synopsis: item.synopsis,
        status: item.status,
        rating: item.rating,
        review: item.review,
        favorite: item.favorite,
        progress: item.progress?.toString() ?? "",
        length: item.length?.toString() ?? "",
        started_at: item.started_at?.slice(0, 10) ?? "",
        finished_at: item.finished_at?.slice(0, 10) ?? "",
        times_consumed: item.times_consumed.toString(),
        source: item.source,
        external_id: item.external_id,
    });
    coverOk.value = true;
}

function intOrNull(v: NumberField) {
    if (typeof v === "number") return Number.isFinite(v) ? v : null;
    return v.trim() === "" ? null : Number(v);
}

function toPayload(): MediaInput {
    return {
        type: form.type,
        title: form.title.trim(),
        creator: form.creator,
        release_year: intOrNull(form.release_year),
        genres: form.genres,
        cover_url: form.cover_url.trim(),
        synopsis: form.synopsis,
        status: form.status,
        rating: form.rating,
        review: form.review,
        favorite: form.favorite,
        progress: intOrNull(form.progress),
        length: intOrNull(form.length),
        // Leave dates out when blank on create so the backend can fill them
        // in from the status; on edit, blank means clear.
        started_at: form.started_at || (isNew.value ? undefined : null),
        finished_at: form.finished_at || (isNew.value ? undefined : null),
        times_consumed: intOrNull(form.times_consumed) ?? 0,
        source: form.source,
        external_id: form.external_id,
    };
}

// The form is a draft: fill it once per item, so a background refetch never clobbers unsaved edits.
let filledId: number | null = null;
watch(
    () => itemQuery.data.value,
    (item) => {
        if (isNew.value || !item || item.ID === filledId) return;
        filledId = item.ID;
        toForm(item);
    },
    { immediate: true },
);

function load() {
    void itemQuery.refetch();
}

async function save() {
    if (!form.title.trim() || isSaving.value) return;
    try {
        if (isNew.value) {
            const item = await createMut.mutateAsync(toPayload());
            toast.push("Added", "success");
            filledId = item.ID;
            await router.replace({ name: "media-detail", params: { id: item.ID } });
            toForm(item);
        } else {
            await updateMut.mutateAsync({ id: itemId.value, payload: toPayload() });
            toast.push("Saved", "success");
            await router.push({ name: "media" });
        }
    } catch {
        // The mutation helper already toasted the error; the form keeps its inputs.
    }
}

/** Quick status change: the form and caches update now, the server fills in dates and we adopt them. */
function quickStatus(status: MediaStatus) {
    const previous = form.status;
    form.status = status;
    if (isNew.value) return;
    quickMut.mutate(
        { id: itemId.value, fields: { status } },
        {
            onSuccess: (item) => {
                // Adopt server-filled dates (and only those) so unsaved edits elsewhere survive.
                if (form.status === item.status) {
                    form.started_at = item.started_at?.slice(0, 10) ?? "";
                    form.finished_at = item.finished_at?.slice(0, 10) ?? "";
                    form.progress = item.progress?.toString() ?? "";
                }
            },
            onError: () => {
                if (form.status === status) form.status = previous;
            },
        },
    );
}

async function remove() {
    const ok = await confirm({
        title: "Delete this item?",
        tone: "danger",
        message: `"${form.title}" and your review will be removed.`,
        confirmText: "Delete",
    });
    if (!ok) return;
    try {
        await deleteMut.mutateAsync(itemId.value);
        toast.push("Deleted", "success");
        await router.push({ name: "media" });
    } catch {
        // Toasted by the mutation helper.
    }
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-4xl flex-col gap-6">
        <PageHeader :title="isNew ? 'Add media' : form.title || 'Media'" />

        <ErrorState v-if="hasError" :error="loadError" title="Couldn't load this item" @retry="load" />
        <div v-else-if="isLoading" class="grid grid-cols-1 gap-6 lg:grid-cols-[200px_1fr]">
            <template v-if="showSkeleton">
                <div class="aspect-[2/3] w-full max-w-50"><Skeleton shape="block" width="100%" height="100%" /></div>
                <div class="flex flex-col gap-4">
                    <Skeleton shape="block" height="2.75rem" />
                    <Skeleton shape="block" height="2.75rem" />
                    <Skeleton shape="block" height="8rem" />
                </div>
            </template>
        </div>
        <form v-else class="grid grid-cols-1 gap-6 lg:grid-cols-[200px_1fr]" @submit.prevent="save">
            <aside class="flex flex-col gap-3">
                <div class="aspect-[2/3] w-full max-w-50 overflow-hidden rounded-card border border-line bg-surface-2">
                    <img
                        v-if="form.cover_url && coverOk"
                        :src="form.cover_url"
                        :alt="form.title"
                        referrerpolicy="no-referrer"
                        class="h-full w-full object-cover"
                        @error="coverOk = false"
                    />
                    <div v-else class="flex h-full items-center justify-center p-3 text-center text-sm text-fg-muted">
                        {{ form.cover_url ? "Couldn't load cover" : "No cover" }}
                    </div>
                </div>
                <Input v-model="form.cover_url" type="url" inputmode="url" placeholder="Cover image URL" aria-label="Cover image URL" @input="coverOk = true" />
                <Checkbox v-model="form.favorite" label="Favourite" />
            </aside>

            <div class="flex min-w-0 flex-col gap-6">
                <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_auto_auto]">
                    <Field label="Title" required>
                        <Input v-model="form.title" required placeholder="Title" />
                    </Field>
                    <Field label="Type">
                        <Select v-model="form.type" :options="typeOptions" @update:model-value="closeLookup" />
                    </Field>
                    <div class="flex items-end">
                        <Button
                            block
                            :loading="isLookingUp"
                            :disabled="!form.title.trim() || !canLookup"
                            :title="canLookup ? 'Search online for this title and fill in the details' : `No lookup source for ${MEDIA_TYPE_LABELS[form.type]} yet`"
                            @click="runLookup"
                        >
                            {{ isLookingUp ? "Looking up…" : "Look up" }}
                        </Button>
                    </div>
                </div>

                <Card v-if="lookupResults" padding="none">
                    <div class="flex items-center justify-between gap-2 px-3 py-2 lg:px-4">
                        <h2 class="m-0 text-sm font-medium text-fg-muted">
                            {{ lookupError ? "Lookup failed" : lookupResults.length ? "Pick a match" : "No matches found" }}
                        </h2>
                        <Button variant="ghost" size="sm" @click="closeLookup">Close</Button>
                    </div>
                    <ErrorState v-if="lookupError" compact :error="lookupError" title="Lookup failed" @retry="runLookup" />
                    <EmptyState v-else-if="!lookupResults.length" title="No matches" description="Try a different title, or fill the details in by hand." />
                    <List v-else>
                        <ListRow
                            v-for="r in lookupResults"
                            :key="r.source + r.external_id"
                            :disabled="applyingId !== null"
                            :pending="applyingId === r.external_id"
                            @click="applyLookup(r)"
                        >
                            <template #leading>
                                <span class="block aspect-[2/3] w-10 overflow-hidden rounded-control bg-surface-3">
                                    <img v-if="r.cover_url" :src="r.cover_url" alt="" loading="lazy" referrerpolicy="no-referrer" class="h-full w-full object-cover" />
                                </span>
                            </template>
                            {{ r.title }}
                            <template #subtitle>{{ [r.release_year, r.creator].filter(Boolean).join(" · ") || "\u00a0" }}</template>
                            <template v-if="applyingId === r.external_id" #trailing><Spinner size="sm" label="Applying" /></template>
                        </ListRow>
                    </List>
                </Card>

                <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_7rem_1fr]">
                    <Field label="Creator"><Input v-model="form.creator" placeholder="Director, author…" /></Field>
                    <Field label="Year"><Input v-model="form.release_year" type="number" inputmode="numeric" placeholder="Year" /></Field>
                    <Field label="Genres"><Input v-model="form.genres" placeholder="Comma separated" /></Field>
                </div>

                <Field label="Status">
                    <!-- Six options don't fit a segmented control at phone width: native picker below lg. -->
                    <div class="lg:hidden">
                        <Select :model-value="form.status" :options="statusOptions" @update:model-value="(v) => quickStatus(v as MediaStatus)" />
                    </div>
                    <div class="hidden lg:block">
                        <SegmentedControl :model-value="form.status" :options="statusOptions" label="Status" @update:model-value="quickStatus" />
                    </div>
                </Field>

                <Field label="Your rating">
                    <div class="flex items-center gap-3">
                        <StarRating v-model="form.rating" />
                        <span class="text-sm text-fg-muted">{{ form.rating != null ? form.rating.toFixed(1) : "Unrated" }}</span>
                    </div>
                </Field>

                <Field label="Thoughts">
                    <Textarea v-model="form.review" :rows="5" placeholder="What did you think?" />
                </Field>

                <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                    <Field :label="`Progress (${unit})`">
                        <div class="flex items-center gap-2">
                            <div class="min-w-0 flex-1"><Input v-model="form.progress" type="number" inputmode="numeric" aria-label="Progress" /></div>
                            <span class="text-fg-muted">/</span>
                            <div class="min-w-0 flex-1"><Input v-model="form.length" type="number" inputmode="numeric" placeholder="total" aria-label="Total" /></div>
                        </div>
                    </Field>
                    <Field label="Times finished"><Input v-model="form.times_consumed" type="number" inputmode="numeric" /></Field>
                    <Field label="Started"><Input v-model="form.started_at" type="date" /></Field>
                    <Field label="Finished"><Input v-model="form.finished_at" type="date" /></Field>
                </div>

                <Field label="Synopsis">
                    <Textarea v-model="form.synopsis" :rows="3" placeholder="What's it about?" />
                </Field>

                <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-between">
                    <Button v-if="!isNew" variant="danger" :loading="isDeleting" @click="remove">Delete</Button>
                    <span v-else />
                    <div class="flex flex-col gap-2 lg:flex-row">
                        <Button :to="{ name: 'media' }">Back</Button>
                        <Button type="submit" variant="primary" :loading="isSaving" :disabled="!form.title.trim()">
                            {{ isNew ? "Add" : "Save" }}
                        </Button>
                    </div>
                </div>
            </div>
        </form>
    </div>
</template>
