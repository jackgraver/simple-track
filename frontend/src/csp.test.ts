import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

// The production CSP (deploy/Caddyfile) allows inline scripts only by hash. Vite copies the inline
// scripts in index.html into the build unchanged, so each one's hash must be in the policy, or the
// browser blocks it (the theme script: wrong theme until the app boots).
// Changing an inline script: put the new hash in deploy/Caddyfile and deploy Caddy before the app.
const read = (rel: string) => readFileSync(fileURLToPath(new URL(rel, import.meta.url)), "utf8");

describe("Content-Security-Policy", () => {
    const html = read("../index.html");
    const caddyfile = read("../../deploy/Caddyfile");
    const csp = /Content-Security-Policy "([^"]+)"/.exec(caddyfile)?.[1] ?? "";
    const scriptSrc = csp.split(";").map((d) => d.trim()).find((d) => d.startsWith("script-src ")) ?? "";

    it("is set in the Caddyfile", () => {
        expect(scriptSrc).not.toBe("");
    });

    const inline = [...html.matchAll(/<script(?![^>]*\bsrc=)[^>]*>([\s\S]*?)<\/script>/g)].map((m) => m[1]);
    it.each(inline.map((body, i) => [i, body] as const))("allows inline script %i by hash", (_, body) => {
        const hash = `'sha256-${createHash("sha256").update(body, "utf8").digest("base64")}'`;
        expect(scriptSrc.split(/\s+/)).toContain(hash);
    });
});
