import type { Transition } from "motion-v";

// Motion tokens for motion-v, the JS counterpart of the duration utilities in docs/design/TOKENS.md.
// Springs with no bounce: they settle like ease-out but carry a swipe's velocity into the animation.

/** Week carousel: the slide between weeks and the selected-day outline (owner-picked 350ms). */
export const carouselSpring: Transition = { type: "spring", visualDuration: 0.35, bounce: 0 };

/** Swipe distance (share of the strip's width) or speed (px/s) that commits to the next week. */
export const SWIPE_DISTANCE = 0.25;
export const SWIPE_VELOCITY = 400;

// Respect the OS reduced-motion setting for JS animations too (the CSS rule in style.css
// only clamps CSS transitions). Mounted once in App.vue.
export { MotionConfig } from "motion-v";
