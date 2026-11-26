# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

One person: the owner, who builds and uses it. Use is split across devices by module. Gym logging happens almost entirely on a phone, mid-workout. Other modules (diet planning, finances, media) are mostly desktop or split between phone and desktop. Multi-user auth exists in the backend, but no other audience is confirmed.

## Product Purpose

A personal tracker for gym, diet, money, media, grocery and daily body metrics (weight, water, steps). It exists because no off-the-shelf app worked: the goal is to click as few buttons as possible and have it keep track of all the numbers. Success is logging something in a few taps and trusting that it is recorded.

## Positioning

Fewest taps, fast and simple, in one app that covers every personal-tracking domain the owner cares about. Self-hosted, private and personalized. It is also a personal engineering project, so it is shaped to the owner's own habits rather than a mass audience. A generic tracker could not truthfully claim any of that.

## Operating Context

- Go API (gin + gorm, Postgres) with a Vue 3 + Vite frontend, deployed to the owner's VPS via j-cicd.
- Gym: logging sets and exercises on a phone during a workout (mobile-first, bottom tab bar, focus pages).
- Diet, finances and media: planning and review, often at a desktop (side nav).
- Daily metrics: quick entry of weight, water and steps.

## Capabilities and Constraints

- Modules (frontend pages): auth, gym, diet, finances, grocery, home, media, settings. Backend also has tracking (weight, water, steps, profile, missed, grocery).
- Theme: dark by default, light supported.
- Logging must feel instant and never blocked on the network. Optimistic UI is a tracked project (`docs/projects/optimistic-ui`).
- The current UI overhaul is about consistency and function, not new features or a new brand. Future features are expected later and the design must leave room for them.
- Long-running work is tracked in `docs/projects/` (ui-overhaul, optimistic-ui, performance).

## Brand Commitments

None established beyond the name "Simple Tracking" and the existing red accent (`cfRed` token). The overhaul explicitly does not introduce a new brand.

## Evidence on Hand

- Real dev seed data (`backend/cmd/devseed`) for gym, diet, tracking, media and money.
- Audit notes in `docs/projects/ui-overhaul/README.md`.
- No testimonials, customers or metrics. Do not fabricate any.

## Product Principles

1. Fewest taps wins. Every flow is judged by how many taps and how much attention it costs.
2. Instant and trustworthy. Logging responds immediately, and pending or failed state is visible, never silent.
3. Fit the device to the job. Gym is a phone-first, one-handed context. Planning and review get desktop density. Neither is a stretched copy of the other.
4. Quiet and functional. The tool recedes, and data and the next action lead.
5. Personal but extensible. It is built for one owner's habits, and new modules must slot in without redesign.

## Accessibility & Inclusion

Baseline WCAG AA: contrast, touch target size, and respect for reduced motion. Dark and light themes both need to meet it.
