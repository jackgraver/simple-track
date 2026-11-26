import type { Transition } from "motion-v";
import { tunable } from "./tune";

// Motion tokens for motion-v, the JS counterpart of the duration utilities in docs/design/TOKENS.md.
// Springs with no bounce: they settle like ease-out but carry a swipe's velocity into the animation.
// Each knob's `value` is the shipped default; `?tune` in dev opens live sliders for them (tune.ts).

/** Week carousel: the slide between weeks and the selected-day outline (owner-picked 350ms), and the
 * swipe distance (share of the strip's width) or speed (px/s) that commits to the next week. */
export const carousel = tunable("Week carousel", {
    duration: { value: 0.35, min: 0.1, max: 1.2, step: 0.01 },
    bounce: { value: 0, min: 0, max: 0.5, step: 0.01 },
    swipeDistance: { value: 0.25, min: 0.05, max: 0.6, step: 0.01 },
    swipeVelocity: { value: 400, min: 50, max: 2000, step: 10 },
});

// Imperative animate() calls don't see <MotionConfig>, so springs check the OS setting themselves.
const reducedQuery = typeof matchMedia === "function" ? matchMedia("(prefers-reduced-motion: reduce)") : null;
export const reduceMotion = (): boolean => reducedQuery?.matches ?? false;

/** Build springs at the moment of animating, so a tuned value applies to the next animation. */
const spring = (duration: number, bounce: number, velocity = 0): Transition =>
    reduceMotion() ? { duration: 0 } : { type: "spring", visualDuration: duration, bounce, velocity };

export const carouselSpring = (): Transition => spring(carousel.duration, carousel.bounce);

/** Mobile nav drawer: the slide in/out (carrying a release's velocity, px/s), and when a released
 * swipe opens it: a fling faster than `flingVelocity` px/s goes its way, otherwise it opens if more
 * than `openShare` of the panel is showing. */
export const drawer = tunable("Nav drawer", {
    duration: { value: 0.3, min: 0.1, max: 1, step: 0.01 },
    bounce: { value: 0, min: 0, max: 0.5, step: 0.01 },
    flingVelocity: { value: 400, min: 100, max: 2000, step: 10 },
    openShare: { value: 0.5, min: 0.2, max: 0.8, step: 0.01 },
});
export const drawerSpring = (velocity = 0): Transition => spring(drawer.duration, drawer.bounce, velocity);

/** Sliding highlight (desktop nav): how fast the hover and active pills move, and the press shrink. */
export const highlight = tunable("Nav highlight", {
    duration: { value: 0.15, min: 0.05, max: 0.6, step: 0.01 },
    pressScale: { value: 0.97, min: 0.85, max: 1, step: 0.005 },
    pressDuration: { value: 0.1, min: 0.03, max: 0.4, step: 0.01 },
});
export const highlightSpring = (): Transition => spring(highlight.duration, 0);
export const pressTransition = (): Transition =>
    reduceMotion() ? { duration: 0 } : { duration: highlight.pressDuration, ease: "easeOut" };

// Respect the OS reduced-motion setting for JS animations too (the CSS rule in style.css
// only clamps CSS transitions). Mounted once in App.vue.
export { MotionConfig } from "motion-v";
