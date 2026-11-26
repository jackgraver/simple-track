-- Unilateral exercise logging: an exercise can be logged one side at a time.
-- exercises.unilateral turns it on per exercise (default off).
-- logged_sets.side is 'left' or 'right' for a set logged on one side, NULL for a bilateral set
-- (every set before this migration). A unilateral set number is a pair of rows, one per side.

-- migrate:up
ALTER TABLE public.exercises ADD COLUMN unilateral boolean NOT NULL DEFAULT false;
ALTER TABLE public.logged_sets ADD COLUMN side text;
ALTER TABLE public.logged_sets ADD CONSTRAINT logged_sets_side_check CHECK (side IN ('left', 'right'));

-- migrate:down
ALTER TABLE public.logged_sets DROP CONSTRAINT logged_sets_side_check;
ALTER TABLE public.logged_sets DROP COLUMN side;
ALTER TABLE public.exercises DROP COLUMN unilateral;
