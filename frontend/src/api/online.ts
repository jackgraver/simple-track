/**
 * Connection state as TanStack Query sees it. While offline, mutations are paused, not failed: their
 * optimistic patch is on screen and the request goes out when the browser reports it is back online
 * (an in-memory queue; a reload drops it). UIs use this to say "waiting for connection" instead of
 * showing a pending row that never resolves.
 */
import { onlineManager } from "@tanstack/vue-query";
import { onScopeDispose, ref, type Ref } from "vue";

export function useOnline(): Readonly<Ref<boolean>> {
    const online = ref(onlineManager.isOnline());
    const unsubscribe = onlineManager.subscribe((value) => (online.value = value));
    onScopeDispose(unsubscribe);
    return online;
}
