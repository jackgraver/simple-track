-- Plates first: each exercise gets a base weight (bar, sled or starting weight; NULL = the load
-- type's default) and a weight step (NULL = the load type's default), and each logged set keeps a
-- snapshot of the plates it was lifted with ({"base": 45, "mode": "per_side", "plates": {"45": 1}}).
-- plate_loaded_with_bar / plate_loaded_without_bar merge into plate_loaded_per_side, with the bar
-- carried over as base_weight 45 / 0.
--
-- Down maps per-side exercises back by base weight (0 = without bar, anything else = with bar),
-- so a custom base weight such as a 25 lb EZ bar comes back as the 45 lb default.

-- migrate:up
ALTER TABLE public.exercises
    ADD COLUMN base_weight numeric,
    ADD COLUMN weight_step numeric;

UPDATE public.exercises
SET load_type = 'plate_loaded_per_side', base_weight = 45
WHERE load_type = 'plate_loaded_with_bar';

UPDATE public.exercises
SET load_type = 'plate_loaded_per_side', base_weight = 0
WHERE load_type = 'plate_loaded_without_bar';

ALTER TABLE public.exercises ALTER COLUMN load_type SET DEFAULT 'plate_loaded_per_side';

ALTER TABLE public.logged_sets ADD COLUMN plate_setup jsonb;

-- migrate:down
ALTER TABLE public.logged_sets DROP COLUMN plate_setup;

ALTER TABLE public.exercises ALTER COLUMN load_type SET DEFAULT 'plate_loaded_with_bar';

UPDATE public.exercises
SET load_type = CASE WHEN base_weight = 0 THEN 'plate_loaded_without_bar' ELSE 'plate_loaded_with_bar' END
WHERE load_type = 'plate_loaded_per_side';

ALTER TABLE public.exercises
    DROP COLUMN base_weight,
    DROP COLUMN weight_step;
