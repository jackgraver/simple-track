import type { ComputedRef, Ref } from "vue";
import { watch } from "vue";

/** Passed into `tryRestore` so callers can delete the key (e.g. invalid payload). */
export type WebStorageJsonSyncContext = {
    remove: () => void;
};

/**
 * Syncs a plain object snapshot to `localStorage` or `sessionStorage` under a reactive key.
 * Watches refs and persists after updates; restore is explicit. Use `setSaveEnabled(false)`
 * until initial state is loaded so the first watch does not overwrite storage.
 */
export function useWebStorageJsonSync<T extends object>(options: {
    /** Defaults to `localStorage`; use `sessionStorage` for tab-only scope. */
    storage?: Storage;
    /** Storage key; re-read whenever this changes (e.g. route/day/entity). */
    key: ComputedRef<string>;
    /** Refs that should trigger a save when any of them change. */
    watchSources: Ref[];
    getSnapshot: () => T;
    /** Apply parsed JSON to app state; return false if ignored. Use `ctx.remove()` to drop bad entries. */
    tryRestore: (
        parsed: Partial<T>,
        ctx: WebStorageJsonSyncContext,
    ) => boolean;
    /** If false, skips save and restore (e.g. missing entity id). */
    canPersist?: () => boolean;
}) {
    const storage = options.storage ?? window.localStorage;
    let saveEnabled = false;

    /** Removes the current key from storage. */
    const remove = () => {
        const k = options.key.value;
        if (k) storage.removeItem(k);
    };

    /** Writes `getSnapshot()` JSON under the current key (no-op if save disabled or cannot persist). */
    const save = () => {
        if (!saveEnabled) return;
        if (options.canPersist?.() === false) return;
        const k = options.key.value;
        if (!k) return;
        storage.setItem(k, JSON.stringify(options.getSnapshot()));
    };

    /** Reads JSON, runs `tryRestore`, or removes the key on parse failure. Returns whether restore applied. */
    const restore = (): boolean => {
        if (options.canPersist?.() === false) return false;
        const k = options.key.value;
        if (!k) return false;
        const raw = storage.getItem(k);
        if (!raw) return false;
        try {
            const parsed = JSON.parse(raw) as Partial<T>;
            return options.tryRestore(parsed, { remove });
        } catch {
            remove();
            return false;
        }
    };

    /** When false, `save` is a no-op (watch still runs but does nothing useful until enabled). */
    const setSaveEnabled = (v: boolean) => {
        saveEnabled = v;
    };

    watch(options.watchSources, save, { flush: "post" });

    return {
        save,
        restore,
        clear: remove,
        setSaveEnabled,
    };
}
