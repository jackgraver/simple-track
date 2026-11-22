/**
 * Client-generated ids for rows created optimistically (DECISIONS.md: bigint PKs + `client_id uuid`).
 *
 * `crypto.randomUUID()` only exists in secure contexts, so it is undefined when the app is opened over
 * plain http on the LAN (testing from a phone). Always call `newClientId()` instead.
 */

type CryptoLike = {
    randomUUID?: () => string;
    getRandomValues?: <T extends ArrayBufferView | null>(array: T) => T;
};

function randomBytes(c: CryptoLike | undefined): Uint8Array {
    const bytes = new Uint8Array(16);
    // getRandomValues is available in insecure contexts too; Math.random is the last resort (ids only
    // need to be unique per user, not unguessable).
    if (typeof c?.getRandomValues === "function") return c.getRandomValues(bytes);
    for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256);
    return bytes;
}

/** RFC 4122 version-4 UUID built from 16 random bytes. */
function uuidV4From(bytes: Uint8Array): string {
    bytes[6] = (bytes[6] & 0x0f) | 0x40; // version 4
    bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant 10xx
    const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

/** A new v4 UUID for a row's `client_id`. Works without `crypto.randomUUID` (plain-http LAN). */
export function newClientId(): string {
    const c = (globalThis as { crypto?: CryptoLike }).crypto;
    if (typeof c?.randomUUID === "function") return c.randomUUID();
    return uuidV4From(randomBytes(c));
}
