-- Diet write safety for optimistic UI (optimistic-ui D1).
-- * day_logs: at most one live row per (day, meal), so a double tap cannot log a meal twice.
--   Existing duplicates are soft-deleted first (oldest row wins).
-- * meals / planned_meals: client_id (client-generated uuid) makes creates idempotent on retry.
--   NULL for rows created before this migration or by clients that do not send one.

-- migrate:up
UPDATE public.day_logs d
SET deleted_at = now()
WHERE d.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM public.day_logs k
    WHERE k.day_id = d.day_id AND k.meal_id = d.meal_id
      AND k.deleted_at IS NULL AND k.id < d.id
  );

CREATE UNIQUE INDEX uq_day_logs_day_meal ON public.day_logs USING btree (day_id, meal_id) WHERE deleted_at IS NULL;

ALTER TABLE public.meals ADD COLUMN client_id uuid;
CREATE UNIQUE INDEX uq_meals_client_id ON public.meals USING btree (client_id) WHERE client_id IS NOT NULL;

ALTER TABLE public.planned_meals ADD COLUMN client_id uuid;
CREATE UNIQUE INDEX uq_planned_meals_client_id ON public.planned_meals USING btree (client_id) WHERE client_id IS NOT NULL;

-- migrate:down
DROP INDEX public.uq_planned_meals_client_id;
ALTER TABLE public.planned_meals DROP COLUMN client_id;
DROP INDEX public.uq_meals_client_id;
ALTER TABLE public.meals DROP COLUMN client_id;
DROP INDEX public.uq_day_logs_day_meal;
