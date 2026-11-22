-- Safe set writes for optimistic gym logging (optimistic-ui G1).
--
-- 1. Merge duplicate live logged_exercises rows for the same (workout_log_id, exercise_id).
--    Concurrent first-set POSTs used to create two parents. Sets move to the lowest id,
--    non-empty notes are concatenated onto it, and the other rows are deleted.
-- 2. One live logged_exercise per (workout_log_id, exercise_id), so the set endpoint can
--    upsert its parent with ON CONFLICT.
-- 3. Missing FK indexes on logged_exercises.workout_log_id and logged_sets.logged_exercise_id.
-- 4. logged_sets.client_id: a client-generated uuid that makes POST /workout/sets
--    idempotent. Rows created before this migration keep NULL (allowed many times).
--
-- The down section drops the indexes and the column. The merge in step 1 is not undone.

-- migrate:up
CREATE TEMP TABLE _logged_exercise_dupes ON COMMIT DROP AS
SELECT id, keep_id
FROM (
    SELECT id, min(id) OVER (PARTITION BY workout_log_id, exercise_id) AS keep_id
    FROM public.logged_exercises
    WHERE deleted_at IS NULL AND workout_log_id IS NOT NULL AND exercise_id IS NOT NULL
) ranked
WHERE id <> keep_id;

UPDATE public.logged_sets s
SET logged_exercise_id = d.keep_id
FROM _logged_exercise_dupes d
WHERE s.logged_exercise_id = d.id;

UPDATE public.logged_exercises keep
SET notes = merged.notes
FROM (
    SELECT d.keep_id, string_agg(le.notes, E'\n' ORDER BY le.id) AS notes
    FROM (SELECT DISTINCT keep_id FROM _logged_exercise_dupes) d
    JOIN public.logged_exercises le
      ON le.id = d.keep_id OR le.id IN (SELECT id FROM _logged_exercise_dupes x WHERE x.keep_id = d.keep_id)
    WHERE coalesce(le.notes, '') <> ''
    GROUP BY d.keep_id
) merged
WHERE keep.id = merged.keep_id;

DELETE FROM public.logged_exercises le
USING _logged_exercise_dupes d
WHERE le.id = d.id;

CREATE UNIQUE INDEX idx_logged_exercises_workout_exercise
    ON public.logged_exercises USING btree (workout_log_id, exercise_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_logged_exercises_workout_log_id ON public.logged_exercises USING btree (workout_log_id);

CREATE INDEX idx_logged_sets_logged_exercise_id ON public.logged_sets USING btree (logged_exercise_id);

ALTER TABLE public.logged_sets ADD COLUMN client_id uuid;

CREATE UNIQUE INDEX idx_logged_sets_client_id ON public.logged_sets USING btree (client_id);

-- migrate:down
DROP INDEX public.idx_logged_sets_client_id;
ALTER TABLE public.logged_sets DROP COLUMN client_id;
DROP INDEX public.idx_logged_sets_logged_exercise_id;
DROP INDEX public.idx_logged_exercises_workout_log_id;
DROP INDEX public.idx_logged_exercises_workout_exercise;
