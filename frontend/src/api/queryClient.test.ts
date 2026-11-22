import { AxiosError, AxiosHeaders } from "axios";
import { describe, expect, it, vi } from "vitest";

import { createQueryClient, isClientError, shouldRetryQuery } from "./queryClient";

function httpError(status: number): AxiosError {
    const config = { headers: new AxiosHeaders() };
    return new AxiosError("fail", "ERR_BAD_REQUEST", config, null, {
        status,
        statusText: "",
        headers: {},
        config,
        data: {},
    });
}

describe("isClientError", () => {
    it("is true for 4xx only", () => {
        for (const s of [400, 401, 404, 409, 499]) expect(isClientError(httpError(s))).toBe(true);
        for (const s of [200, 399, 500, 503]) expect(isClientError(httpError(s))).toBe(false);
    });
    it("is false for network and non-axios errors", () => {
        expect(isClientError(new AxiosError("Network Error"))).toBe(false);
        expect(isClientError(new Error("boom"))).toBe(false);
        expect(isClientError(undefined)).toBe(false);
    });
});

describe("shouldRetryQuery", () => {
    it("retries 5xx and network errors up to twice", () => {
        const err = httpError(500);
        expect(shouldRetryQuery(0, err)).toBe(true);
        expect(shouldRetryQuery(1, err)).toBe(true);
        expect(shouldRetryQuery(2, err)).toBe(false);
        expect(shouldRetryQuery(0, new AxiosError("Network Error"))).toBe(true);
    });
    it("never retries 4xx", () => {
        expect(shouldRetryQuery(0, httpError(404))).toBe(false);
    });
});

describe("createQueryClient", () => {
    it("applies the DESIGN.md defaults", () => {
        const d = createQueryClient(vi.fn()).getDefaultOptions();
        expect(d.queries?.staleTime).toBe(60_000);
        expect(d.queries?.gcTime).toBe(24 * 60 * 60_000);
        expect(d.queries?.refetchOnWindowFocus).toBe(true);
        expect(d.queries?.retry).toBe(shouldRetryQuery);
        expect(d.mutations?.retry).toBe(0);
    });

    it("a 404 query fails after a single request", async () => {
        const qc = createQueryClient(vi.fn());
        const fn = vi.fn().mockRejectedValue(httpError(404));
        await expect(qc.fetchQuery({ queryKey: ["x"], queryFn: fn, retryDelay: 0 })).rejects.toBeTruthy();
        expect(fn).toHaveBeenCalledTimes(1);
    });

    it("toasts a failed refetch only when cached data exists and not silent", async () => {
        const notify = vi.fn();
        const qc = createQueryClient(notify);
        qc.setQueryData(["a"], 1);
        await qc.fetchQuery({ queryKey: ["a"], queryFn: () => Promise.reject(httpError(500)), staleTime: 0, retry: false }).catch(() => {});
        expect(notify).toHaveBeenCalledWith("Couldn’t refresh");

        notify.mockClear();
        await qc.fetchQuery({ queryKey: ["cold"], queryFn: () => Promise.reject(httpError(500)), retry: false }).catch(() => {});
        qc.setQueryData(["s"], 1);
        await qc.fetchQuery({ queryKey: ["s"], queryFn: () => Promise.reject(httpError(500)), staleTime: 0, retry: false, meta: { silent: true } }).catch(() => {});
        expect(notify).not.toHaveBeenCalled();
    });

    it("toasts mutation errors unless handled or 401", async () => {
        const notify = vi.fn();
        const qc = createQueryClient(notify);
        const run = (err: unknown, meta?: Record<string, unknown>) =>
            qc.getMutationCache().build(qc, { mutationFn: () => Promise.reject(err), meta }).execute(undefined).catch(() => {});
        await run(httpError(500));
        expect(notify).toHaveBeenCalledTimes(1);
        await run(httpError(500), { handled: true });
        await run(httpError(401));
        expect(notify).toHaveBeenCalledTimes(1);
    });
});
