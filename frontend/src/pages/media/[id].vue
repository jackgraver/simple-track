<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
    createMedia,
    deleteMedia,
    getMedia,
    LOOKUP_SOURCE_LABELS,
    LOOKUP_TYPES,
    lookupMedia,
    lookupMediaDetails,
    MEDIA_STATUS_LABELS,
    MEDIA_STATUSES,
    MEDIA_TYPE_LABELS,
    MEDIA_TYPES,
    updateMedia,
    type LookupResult,
    type MediaInput,
    type MediaItem,
    type MediaStatus,
    type MediaType,
} from "~/api/media/api";
import { getApiErrorMessage } from "~/api/httpError";
import { useDialog } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import StarRating from "./StarRating.vue";

const route = useRoute();
const router = useRouter();
const { confirm } = useDialog();

const isNew = computed(() => route.name === "media-new");
const itemId = computed(() => Number(route.params.id));

const isLoading = ref(!isNew.value);
const isSaving = ref(false);
const errorMessage = ref("");

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

async function load() {
    if (isNew.value) return;
    isLoading.value = true;
    errorMessage.value = "";
    try {
        toForm(await getMedia(itemId.value));
    } catch (error) {
        errorMessage.value = getApiErrorMessage(error, "Failed to load item");
    } finally {
        isLoading.value = false;
    }
}

async function save() {
    if (!form.title.trim() || isSaving.value) return;
    isSaving.value = true;
    try {
        if (isNew.value) {
            const item = await createMedia(toPayload());
            toast.push("Added", "success");
            await router.replace({ name: "media-detail", params: { id: item.ID } });
            toForm(item);
        } else {
            await updateMedia(itemId.value, toPayload());
            toast.push("Saved", "success");
            await router.push({ name: "media" });
        }
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to save"), "error");
    } finally {
        isSaving.value = false;
    }
}

/** Quick status change that saves immediately (backend fills in dates). */
async function quickStatus(status: MediaStatus) {
    form.status = status;
    if (isNew.value) return;
    isSaving.value = true;
    try {
        toForm(await updateMedia(itemId.value, { status }));
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to update status"), "error");
    } finally {
        isSaving.value = false;
    }
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
        await deleteMedia(itemId.value);
        toast.push("Deleted", "success");
        await router.push({ name: "media" });
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to delete"), "error");
    }
}

onMounted(() => void load());
</script>

<template>
    <div class="mx-auto flex w-full max-w-4xl flex-col gap-6">
        <p v-if="errorMessage" class="m-0 text-sm text-(--color-cf-red)">{{ errorMessage }}</p>
        <p v-else-if="isLoading" class="m-0 text-sm text-textSecondary">Loading…</p>
        <form v-else class="grid gap-6 md:grid-cols-[200px_1fr]" @submit.prevent="save">
            <aside class="flex flex-col gap-3">
                <div class="aspect-[2/3] w-full max-w-50 overflow-hidden rounded-lg bg-secondBg">
                    <img
                        v-if="form.cover_url && coverOk"
                        :src="form.cover_url"
                        :alt="form.title"
                        referrerpolicy="no-referrer"
                        class="h-full w-full object-cover"
                        @error="coverOk = false"
                    />
                    <div v-else class="flex h-full items-center justify-center p-3 text-center text-sm text-textSecondary">
                        {{ form.cover_url ? "Couldn't load cover" : "No cover" }}
                    </div>
                </div>
                <input
                    v-model="form.cover_url"
                    type="url"
                    placeholder="Cover image URL"
                    class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none"
                    @input="coverOk = true"
                />
                <label class="flex items-center gap-2 text-sm">
                    <input v-model="form.favorite" type="checkbox" /> Favourite
                </label>
            </aside>

            <div class="flex min-w-0 flex-col gap-5">
                <div class="grid grid-cols-[1fr_auto] gap-2 sm:grid-cols-[1fr_auto_auto]">
                    <input
                        v-model="form.title"
                        required
                        placeholder="Title"
                        class="col-span-2 rounded-md bg-thirdBg px-3 py-2 text-lg font-semibold outline-none sm:col-span-1"
                    />
                    <select v-model="form.type" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" @change="closeLookup">
                        <option v-for="t in MEDIA_TYPES" :key="t" :value="t">{{ MEDIA_TYPE_LABELS[t] }}</option>
                    </select>
                    <button
                        type="button"
                        class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg disabled:opacity-50"
                        :disabled="!form.title.trim() || !canLookup || isLookingUp"
                        :title="canLookup ? 'Search online for this title and fill in the details' : `No lookup source for ${MEDIA_TYPE_LABELS[form.type]} yet`"
                        @click="runLookup"
                    >
                        {{ isLookingUp ? "Looking up…" : "Look up" }}
                    </button>
                </div>

                <section v-if="lookupResults" class="flex flex-col gap-2 rounded-lg bg-secondBg p-3">
                    <div class="flex items-center justify-between gap-2">
                        <h2 class="m-0 text-sm font-medium text-textSecondary">
                            {{ lookupError ? "Lookup failed" : lookupResults.length ? "Pick a match" : "No matches found" }}
                        </h2>
                        <button type="button" class="text-sm text-textSecondary hover:underline" @click="closeLookup">Close</button>
                    </div>
                    <p v-if="lookupError" class="m-0 text-sm text-(--color-cf-red)">{{ lookupError }}</p>
                    <button
                        v-for="r in lookupResults"
                        :key="r.source + r.external_id"
                        type="button"
                        class="flex items-center gap-3 rounded-md p-1.5 text-left hover:bg-thirdBg disabled:opacity-50"
                        :disabled="applyingId !== null"
                        @click="applyLookup(r)"
                    >
                        <div class="aspect-[2/3] w-10 shrink-0 overflow-hidden rounded bg-thirdBg">
                            <img v-if="r.cover_url" :src="r.cover_url" alt="" loading="lazy" referrerpolicy="no-referrer" class="h-full w-full object-cover" />
                        </div>
                        <div class="min-w-0">
                            <p class="m-0 truncate text-sm font-medium text-textPrimary">{{ r.title }}</p>
                            <p class="m-0 truncate text-xs text-textSecondary">
                                {{ [r.release_year, r.creator].filter(Boolean).join(" · ") || "\u00a0" }}
                            </p>
                        </div>
                        <span v-if="applyingId === r.external_id" class="ml-auto shrink-0 text-xs text-textSecondary">Loading…</span>
                    </button>
                </section>
                <div class="grid gap-2 sm:grid-cols-[1fr_7rem_1fr]">
                    <input v-model="form.creator" placeholder="Creator (director, author…)" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <input v-model="form.release_year" type="number" min="0" max="9999" placeholder="Year" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                    <input v-model="form.genres" placeholder="Genres (comma separated)" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
                </div>

                <section class="flex flex-col gap-2">
                    <h2 class="m-0 text-sm font-medium text-textSecondary">Status</h2>
                    <div class="flex flex-wrap gap-1">
                        <button
                            v-for="s in MEDIA_STATUSES"
                            :key="s"
                            type="button"
                            class="rounded-md px-3 py-1.5 text-sm disabled:opacity-50"
                            :class="form.status === s ? 'bg-primary text-white' : 'bg-secondBg hover:bg-thirdBg'"
                            :disabled="isSaving"
                            @click="quickStatus(s)"
                        >
                            {{ MEDIA_STATUS_LABELS[s] }}
                        </button>
                    </div>
                </section>

                <section class="flex flex-col gap-2">
                    <h2 class="m-0 text-sm font-medium text-textSecondary">Your rating</h2>
                    <div class="flex items-center gap-3">
                        <StarRating v-model="form.rating" />
                        <span class="text-sm text-textSecondary">{{ form.rating != null ? form.rating.toFixed(1) : "Unrated" }}</span>
                    </div>
                </section>

                <section class="flex flex-col gap-2">
                    <h2 class="m-0 text-sm font-medium text-textSecondary">Thoughts</h2>
                    <textarea
                        v-model="form.review"
                        rows="5"
                        placeholder="What did you think?"
                        class="resize-y rounded-md bg-thirdBg px-3 py-2 text-sm outline-none"
                    />
                </section>

                <section class="grid gap-3 sm:grid-cols-2">
                    <label class="flex flex-col gap-1 text-sm text-textSecondary">
                        Progress ({{ unit }})
                        <div class="flex items-center gap-2">
                            <input v-model="form.progress" type="number" min="0" class="w-full rounded-md bg-thirdBg px-3 py-2 text-sm text-textPrimary outline-none" />
                            <span>/</span>
                            <input v-model="form.length" type="number" min="0" placeholder="total" class="w-full rounded-md bg-thirdBg px-3 py-2 text-sm text-textPrimary outline-none" />
                        </div>
                    </label>
                    <label class="flex flex-col gap-1 text-sm text-textSecondary">
                        Times finished
                        <input v-model="form.times_consumed" type="number" min="0" class="rounded-md bg-thirdBg px-3 py-2 text-sm text-textPrimary outline-none" />
                    </label>
                    <label class="flex flex-col gap-1 text-sm text-textSecondary">
                        Started
                        <input v-model="form.started_at" type="date" class="rounded-md bg-thirdBg px-3 py-2 text-sm text-textPrimary outline-none" />
                    </label>
                    <label class="flex flex-col gap-1 text-sm text-textSecondary">
                        Finished
                        <input v-model="form.finished_at" type="date" class="rounded-md bg-thirdBg px-3 py-2 text-sm text-textPrimary outline-none" />
                    </label>
                </section>

                <section class="flex flex-col gap-2">
                    <h2 class="m-0 text-sm font-medium text-textSecondary">Synopsis</h2>
                    <textarea
                        v-model="form.synopsis"
                        rows="3"
                        placeholder="What's it about?"
                        class="resize-y rounded-md bg-thirdBg px-3 py-2 text-sm outline-none"
                    />
                </section>

                <div class="flex flex-wrap justify-between gap-2">
                    <button
                        v-if="!isNew"
                        type="button"
                        class="rounded-md px-3 py-2 text-sm text-(--color-cf-red) hover:bg-secondBg"
                        @click="remove"
                    >
                        Delete
                    </button>
                    <span v-else />
                    <div class="flex gap-2">
                        <router-link :to="{ name: 'media' }" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg">Back</router-link>
                        <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">
                            {{ isNew ? "Add" : "Save" }}
                        </button>
                    </div>
                </div>
            </div>
        </form>
    </div>
</template>
