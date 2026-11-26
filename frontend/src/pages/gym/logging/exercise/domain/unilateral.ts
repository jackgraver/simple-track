import type { SetSide } from "~/types/workout";

/**
 * Unilateral logging: each side is its own logged set (`side` = "left"/"right"), and one set number
 * is a pair of them. Sets without a side are bilateral; on a unilateral exercise (history from before
 * the toggle) one counts for both sides of its pair.
 */

/**
 * `pair_id` ties the two sides of a set number together: the client_id of the side logged first, on
 * both rows. The row's own id is `key` (a today row) or `client_id` (a cached/previous set).
 */
export type SidedSet = {
    weight: number;
    reps: number;
    side?: SetSide | null;
    pair_id?: string | null;
    key?: string;
    client_id?: string | null;
};

/** One logged set inside a pair, with its index in the flat list of rows. */
export type PairEntry<T> = { set: T; index: number };

export type SetPair<T> = {
    left?: PairEntry<T>;
    right?: PairEntry<T>;
    /** A bilateral set: stands for both sides. */
    both?: PairEntry<T>;
    /** The side logged first; null for a bilateral set. */
    first: SetSide | null;
};

export const otherSide = (side: SetSide): SetSide => (side === "left" ? "right" : "left");
export const sideLabel = (side: SetSide): string => (side === "left" ? "Left" : "Right");
export const sideLetter = (side: SetSide): string => (side === "left" ? "L" : "R");

export function hasSides(sets: readonly SidedSet[] | null | undefined): boolean {
    return (sets ?? []).some((s) => s.side === "left" || s.side === "right");
}

const ownId = (set: SidedSet): string | null => set.key ?? set.client_id ?? null;

/**
 * Group rows (in log order) into set numbers, ordered by their first row. A side joins the pair its
 * `pair_id` names (the first side's own id) when that pair doesn't have the side yet, else it starts
 * a new pair; a bilateral set is a pair on its own. Pairing is stored, not derived from order, so
 * deleting or moving one side leaves every other pair where it was. Sides without a `pair_id` (none
 * in production; early dev data) fall back to order: they fill an open pair of the other side.
 */
export function pairSets<T extends SidedSet>(sets: readonly T[] | null | undefined): SetPair<T>[] {
    const pairs: SetPair<T>[] = [];
    const byId = new Map<string, SetPair<T>>();
    const register = (pair: SetPair<T>, ...ids: (string | null | undefined)[]) => {
        for (const id of ids) if (id && !byId.has(id)) byId.set(id, pair);
    };
    (sets ?? []).forEach((set, index) => {
        const side = set.side ?? null;
        if (side == null) {
            pairs.push({ both: { set, index }, first: null });
            return;
        }
        const entry = { set, index };
        const pairId = set.pair_id ?? null;
        if (pairId) {
            const pair = byId.get(pairId);
            if (pair && !pair[side] && !pair.both) {
                pair[side] = entry;
                return;
            }
        } else {
            const last = pairs[pairs.length - 1];
            const lastFirst = last?.first ? last[last.first]?.set : undefined;
            if (last && isOpen(last) && last.first === otherSide(side) && !lastFirst?.pair_id) {
                last[side] = entry;
                return;
            }
        }
        const pair: SetPair<T> = { [side]: entry, first: side };
        pairs.push(pair);
        register(pair, pairId, ownId(set));
    });
    return pairs;
}

/** The id the pair's rows share: the first side's `pair_id`, else its own id. Null if bilateral. */
export function pairIdOf(pair: SetPair<SidedSet> | undefined): string | null {
    const first = pair?.first ? pair[pair.first]?.set : undefined;
    return first ? (first.pair_id ?? ownId(first)) : null;
}

/** The pair a new log on `side` joins (the open last pair, missing that side), if any. */
export function openPairFor<T extends SidedSet>(
    today: readonly T[] | null | undefined,
    side: SetSide,
): SetPair<T> | undefined {
    const pairs = pairSets(today);
    const last = pairs[pairs.length - 1];
    return last && isOpen(last) && last.first === otherSide(side) ? last : undefined;
}

/** Only one side logged so far. */
export function isOpen(pair: SetPair<unknown>): boolean {
    return pair.first != null && !(pair.left && pair.right);
}

/** The set that counts for `side` in a pair (a bilateral set counts for both). */
export function sideSet<T>(pair: SetPair<T> | undefined, side: SetSide): T | undefined {
    return (pair?.[side] ?? pair?.both)?.set;
}

/**
 * One set per pair for progression math (overload, verdict, next weight, planned count): the weaker
 * side, i.e. the lower weight, then the lower reps. A side that dropped weight is "at a lower weight"
 * like a dropped bilateral set, so it doesn't count against or toward the working weight.
 */
export function collapsePair<T extends SidedSet>(pair: SetPair<T>): T {
    if (pair.both) return pair.both.set;
    const sides = [pair.left?.set, pair.right?.set].filter((s): s is T => s != null);
    return sides.reduce((weak, s) =>
        s.weight < weak.weight || (s.weight === weak.weight && s.reps < weak.reps) ? s : weak,
    );
}

/** Sets as progression sees them: pairs collapsed when any set has a side, else unchanged. */
export function setsForProgression<T extends SidedSet>(sets: readonly T[] | null | undefined): T[] {
    if (!sets?.length) return [];
    return hasSides(sets) ? pairSets(sets).map(collapsePair) : [...sets];
}

/** The side to start with: whatever last session started with, else left. */
export function defaultFirstSide(previous: readonly SidedSet[] | null | undefined): SetSide {
    return pairSets(previous).find((p) => p.first != null)?.first ?? "left";
}

/**
 * The side the next log is for: the missing side of an open pair, else the side today's sets start
 * with (or `firstSide` before anything is logged).
 */
export function nextSide(today: readonly SidedSet[] | null | undefined, firstSide: SetSide): SetSide {
    const pairs = pairSets(today);
    const last = pairs[pairs.length - 1];
    if (last && isOpen(last) && last.first) return otherSide(last.first);
    // The side the latest complete set started with (a half-deleted earlier set doesn't count).
    const complete = pairs.filter((p) => p.left && p.right);
    return complete[complete.length - 1]?.first ?? pairs.find((p) => p.first != null)?.first ?? firstSide;
}

/** 1-based set number the next log belongs to: the open pair's, else a new one. */
export function nextPairNumber(today: readonly SidedSet[] | null | undefined): number {
    const pairs = pairSets(today);
    const last = pairs[pairs.length - 1];
    return last && isOpen(last) ? pairs.length : pairs.length + 1;
}

/** Number of complete set numbers (an open pair doesn't count yet). */
export function completePairCount(today: readonly SidedSet[] | null | undefined): number {
    return pairSets(today).filter((p) => !isOpen(p)).length;
}

/** 0-based pair index of the row at `rowIndex`. */
export function pairIndexOfRow(sets: readonly SidedSet[] | null | undefined, rowIndex: number): number {
    return pairSets(sets).findIndex((p) => [p.left, p.right, p.both].some((e) => e?.index === rowIndex));
}

/** Would logging `side` next close an open pair (rather than start a new one)? */
export function closesPair(today: readonly SidedSet[] | null | undefined, side: SetSide): boolean {
    return openPairFor(today, side) != null;
}

/** The latest set logged today on `side`, if any (bilateral sets don't count). */
export function lastOnSide<T extends SidedSet>(today: readonly T[] | null | undefined, side: SetSide): T | undefined {
    const sets = today ?? [];
    for (let i = sets.length - 1; i >= 0; i--) if (sets[i].side === side) return sets[i];
    return undefined;
}

/** "L10 R9" style reps of a pair; uneven weights spell out each side. */
export function formatPair(pair: SetPair<SidedSet>): { weight: number | null; text: string } {
    if (pair.both) return { weight: pair.both.set.weight, text: String(pair.both.set.reps) };
    const parts = (["left", "right"] as const).filter((s) => pair[s]).map((s) => ({ side: s, set: pair[s]!.set }));
    const even = parts.every((p) => p.set.weight === parts[0].set.weight);
    if (even) {
        return { weight: parts[0].set.weight, text: parts.map((p) => `${sideLetter(p.side)}${p.set.reps}`).join(" ") };
    }
    return {
        weight: null,
        text: parts.map((p) => `${sideLetter(p.side)} ${formatWeight(p.set.weight)} × ${p.set.reps}`).join(" "),
    };
}

export function formatWeight(weight: number): string {
    return weight > 0 ? `${weight} lb` : "BW";
}

/**
 * Weight (and setup) for the next log on `side`, when the last complete pair had a different weight
 * on each side: each side keeps its own weight then. Null when the sides matched, so a weight change
 * made on one side (going up for both) carries to the other as usual.
 */
export function carriedSideWeight<T extends SidedSet & { weight_setup?: string }>(
    today: readonly T[] | null | undefined,
    side: SetSide,
): { weight: number; weight_setup: string } | null {
    const complete = pairSets(today).filter((p) => p.left && p.right);
    const ref = complete[complete.length - 1];
    if (!ref?.left || !ref.right || ref.left.set.weight === ref.right.set.weight) return null;
    const set = ref[side]!.set;
    return { weight: set.weight, weight_setup: set.weight_setup ?? "" };
}

/** Last session's set for (0-based) set number `pairIndex` on `side`. */
export function previousSideSet<T extends SidedSet>(
    previous: readonly T[] | null | undefined,
    pairIndex: number,
    side: SetSide,
): T | undefined {
    return sideSet(pairSets(previous)[pairIndex], side);
}
