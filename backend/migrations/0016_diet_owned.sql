-- Diet module (lane C2, docs/multi-user/DECISIONS.md "C2 diet").
--
-- foods is the one shared catalog (created_by, catalog policies, catalog_guard). Other users'
-- quick-entry foods (the hidden row behind a quick-logged meal) are invisible to everyone but
-- their creator. Everything else is owned: plans, days, meals, planned_meals, saved_meals,
-- composite_foods, with owned children day_logs, meal_items, saved_meal_items and
-- composite_food_items (composite FKs on (parent_id, user_id)). All existing rows go to the one
-- admin.

-- migrate:up

-- 1. Columns.
ALTER TABLE public.foods ADD COLUMN created_by bigint;
ALTER TABLE public.plans ADD COLUMN user_id bigint;
ALTER TABLE public.days ADD COLUMN user_id bigint;
ALTER TABLE public.meals ADD COLUMN user_id bigint;
ALTER TABLE public.saved_meals ADD COLUMN user_id bigint;
ALTER TABLE public.composite_foods ADD COLUMN user_id bigint;
ALTER TABLE public.planned_meals ADD COLUMN user_id bigint;
ALTER TABLE public.day_logs ADD COLUMN user_id bigint;
ALTER TABLE public.meal_items ADD COLUMN user_id bigint;
ALTER TABLE public.saved_meal_items ADD COLUMN user_id bigint;
ALTER TABLE public.composite_food_items ADD COLUMN user_id bigint;

-- 2. Backfill: roots to the admin (soft-deleted rows too), children from their parent.
DO $$
DECLARE
  admins bigint;
  owner_id bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM public.foods UNION ALL SELECT 1 FROM public.plans
             UNION ALL SELECT 1 FROM public.days UNION ALL SELECT 1 FROM public.meals
             UNION ALL SELECT 1 FROM public.saved_meals UNION ALL SELECT 1 FROM public.composite_foods) THEN
    SELECT count(*), min(id) INTO admins, owner_id FROM public.users WHERE is_admin AND deleted_at IS NULL;
    IF admins <> 1 THEN
      RAISE EXCEPTION 'diet: % live admin accounts, expected exactly one to own the existing rows', admins;
    END IF;
    UPDATE public.foods SET created_by = owner_id WHERE created_by IS NULL;
    UPDATE public.plans SET user_id = owner_id WHERE user_id IS NULL;
    UPDATE public.days SET user_id = owner_id WHERE user_id IS NULL;
    UPDATE public.meals SET user_id = owner_id WHERE user_id IS NULL;
    UPDATE public.saved_meals SET user_id = owner_id WHERE user_id IS NULL;
    UPDATE public.composite_foods SET user_id = owner_id WHERE user_id IS NULL;
  END IF;
END $$;
UPDATE public.planned_meals c SET user_id = p.user_id FROM public.days p WHERE p.id = c.day_id AND c.user_id IS NULL;
UPDATE public.day_logs c SET user_id = p.user_id FROM public.days p WHERE p.id = c.day_id AND c.user_id IS NULL;
UPDATE public.meal_items c SET user_id = p.user_id FROM public.meals p WHERE p.id = c.meal_id AND c.user_id IS NULL;
UPDATE public.saved_meal_items c SET user_id = p.user_id FROM public.saved_meals p WHERE p.id = c.saved_meal_id AND c.user_id IS NULL;
UPDATE public.composite_food_items c SET user_id = p.user_id FROM public.composite_foods p WHERE p.id = c.composite_food_id AND c.user_id IS NULL;

-- 3. NOT NULL, FK to users, index.
ALTER TABLE public.foods ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE public.foods ADD CONSTRAINT fk_foods_created_by FOREIGN KEY (created_by) REFERENCES public.users (id);
CREATE INDEX idx_foods_created_by ON public.foods USING btree (created_by);

ALTER TABLE public.plans ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.plans ADD CONSTRAINT fk_plans_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_plans_user_id ON public.plans USING btree (user_id);
ALTER TABLE public.days ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.days ADD CONSTRAINT fk_days_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_days_user_id ON public.days USING btree (user_id);
ALTER TABLE public.meals ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.meals ADD CONSTRAINT fk_meals_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_meals_user_id ON public.meals USING btree (user_id);
ALTER TABLE public.saved_meals ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.saved_meals ADD CONSTRAINT fk_saved_meals_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_saved_meals_user_id ON public.saved_meals USING btree (user_id);
ALTER TABLE public.composite_foods ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.composite_foods ADD CONSTRAINT fk_composite_foods_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_composite_foods_user_id ON public.composite_foods USING btree (user_id);

ALTER TABLE public.planned_meals ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_planned_meals_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.day_logs ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.day_logs ADD CONSTRAINT fk_day_logs_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.meal_items ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.meal_items ADD CONSTRAINT fk_meal_items_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.saved_meal_items ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.saved_meal_items ADD CONSTRAINT fk_saved_meal_items_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.composite_food_items ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.composite_food_items ADD CONSTRAINT fk_composite_food_items_user FOREIGN KEY (user_id) REFERENCES public.users (id);

-- 4. (id, user_id) keys for everything that gets pointed at.
ALTER TABLE public.plans ADD CONSTRAINT uq_plans_id_user UNIQUE (id, user_id);
ALTER TABLE public.days ADD CONSTRAINT uq_days_id_user UNIQUE (id, user_id);
ALTER TABLE public.meals ADD CONSTRAINT uq_meals_id_user UNIQUE (id, user_id);
ALTER TABLE public.saved_meals ADD CONSTRAINT uq_saved_meals_id_user UNIQUE (id, user_id);
ALTER TABLE public.composite_foods ADD CONSTRAINT uq_composite_foods_id_user UNIQUE (id, user_id);

-- 5. Composite FKs (FK checks bypass RLS, so a plain FK would accept another user's ID).
-- Existing parent links keep their ON DELETE.
ALTER TABLE public.days DROP CONSTRAINT fk_days_plan;
ALTER TABLE public.days ADD CONSTRAINT fk_days_plan FOREIGN KEY (plan_id, user_id)
  REFERENCES public.plans (id, user_id);

ALTER TABLE public.planned_meals DROP CONSTRAINT fk_days_planned_meals;
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_days_planned_meals FOREIGN KEY (day_id, user_id)
  REFERENCES public.days (id, user_id);
ALTER TABLE public.planned_meals DROP CONSTRAINT fk_planned_meals_meal;
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_planned_meals_meal FOREIGN KEY (meal_id, user_id)
  REFERENCES public.meals (id, user_id);
CREATE INDEX idx_planned_meals_user_day ON public.planned_meals USING btree (user_id, day_id);

ALTER TABLE public.day_logs DROP CONSTRAINT fk_days_logs;
ALTER TABLE public.day_logs ADD CONSTRAINT fk_days_logs FOREIGN KEY (day_id, user_id)
  REFERENCES public.days (id, user_id);
ALTER TABLE public.day_logs DROP CONSTRAINT fk_day_logs_meal;
ALTER TABLE public.day_logs ADD CONSTRAINT fk_day_logs_meal FOREIGN KEY (meal_id, user_id)
  REFERENCES public.meals (id, user_id);
CREATE INDEX idx_day_logs_user_day ON public.day_logs USING btree (user_id, day_id);

ALTER TABLE public.meal_items DROP CONSTRAINT fk_meals_items;
ALTER TABLE public.meal_items ADD CONSTRAINT fk_meals_items FOREIGN KEY (meal_id, user_id)
  REFERENCES public.meals (id, user_id) ON DELETE CASCADE;
CREATE INDEX idx_meal_items_user_meal ON public.meal_items USING btree (user_id, meal_id);

ALTER TABLE public.saved_meal_items DROP CONSTRAINT fk_saved_meals_items;
ALTER TABLE public.saved_meal_items ADD CONSTRAINT fk_saved_meals_items FOREIGN KEY (saved_meal_id, user_id)
  REFERENCES public.saved_meals (id, user_id) ON DELETE CASCADE;
CREATE INDEX idx_saved_meal_items_user_saved_meal ON public.saved_meal_items USING btree (user_id, saved_meal_id);

ALTER TABLE public.composite_food_items DROP CONSTRAINT fk_composite_foods_items;
ALTER TABLE public.composite_food_items ADD CONSTRAINT fk_composite_foods_items FOREIGN KEY (composite_food_id, user_id)
  REFERENCES public.composite_foods (id, user_id) ON DELETE CASCADE;
CREATE INDEX idx_composite_food_items_user_composite ON public.composite_food_items USING btree (user_id, composite_food_id);

-- Columns that pointed at an owned row with no FK at all. NOT VALID enforces new and changed
-- rows without failing on a legacy dangling ID; the VALIDATE below runs where the data is clean
-- and only logs a notice where it is not (ROLLOUT pre-check lists the dangling ones).
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_planned_meals_saved_meal FOREIGN KEY (saved_meal_id, user_id)
  REFERENCES public.saved_meals (id, user_id) NOT VALID;
ALTER TABLE public.meal_items ADD CONSTRAINT fk_meal_items_composite_food FOREIGN KEY (composite_food_id, user_id)
  REFERENCES public.composite_foods (id, user_id) NOT VALID;
ALTER TABLE public.saved_meal_items ADD CONSTRAINT fk_saved_meal_items_composite_food FOREIGN KEY (composite_food_id, user_id)
  REFERENCES public.composite_foods (id, user_id) NOT VALID;
DO $$
DECLARE
  c record;
BEGIN
  FOR c IN SELECT * FROM (VALUES
      ('planned_meals', 'fk_planned_meals_saved_meal'),
      ('meal_items', 'fk_meal_items_composite_food'),
      ('saved_meal_items', 'fk_saved_meal_items_composite_food')) AS t(tbl, con) LOOP
    BEGIN
      EXECUTE format('ALTER TABLE public.%I VALIDATE CONSTRAINT %I', c.tbl, c.con);
    EXCEPTION WHEN foreign_key_violation THEN
      RAISE NOTICE '%.% left NOT VALID: some rows point at a missing parent', c.tbl, c.con;
    END;
  END LOOP;
END $$;

-- 6. Per-user uniques.
DROP INDEX public.idx_days_date;
CREATE UNIQUE INDEX uq_days_user_date ON public.days USING btree (user_id, date);
DROP INDEX public.idx_composite_foods_name;
CREATE UNIQUE INDEX uq_composite_foods_user_name ON public.composite_foods USING btree (user_id, name);
DROP INDEX public.uq_meals_client_id;
CREATE UNIQUE INDEX uq_meals_user_client_id ON public.meals USING btree (user_id, client_id) WHERE client_id IS NOT NULL;
DROP INDEX public.uq_planned_meals_client_id;
CREATE UNIQUE INDEX uq_planned_meals_user_client_id ON public.planned_meals USING btree (user_id, client_id) WHERE client_id IS NOT NULL;

-- 7. Row-level security: owned tables.
ALTER TABLE public.plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.plans USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.days ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.days USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.meals ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.meals USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.saved_meals ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.saved_meals USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.composite_foods ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.composite_foods USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.planned_meals ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.planned_meals USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.day_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.day_logs USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.meal_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.meal_items USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.saved_meal_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.saved_meal_items USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.composite_food_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.composite_food_items USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id());

-- 8. Row-level security: the foods catalog. Everyone signed in reads, except that a quick-entry
-- food (the row behind someone's quick-logged meal) is only visible to its creator.
ALTER TABLE public.foods ENABLE ROW LEVEL SECURITY;
CREATE POLICY catalog_read ON public.foods FOR SELECT
  USING (public.app_user_id() IS NOT NULL
         AND (COALESCE(quick_entry, false) = false OR created_by = public.app_user_id()));
CREATE POLICY catalog_add ON public.foods FOR INSERT
  WITH CHECK (created_by = public.app_user_id());
CREATE POLICY catalog_edit ON public.foods FOR UPDATE
  USING (created_by = public.app_user_id() OR public.app_is_admin())
  WITH CHECK (created_by = public.app_user_id() OR public.app_is_admin());
CREATE POLICY catalog_delete ON public.foods FOR DELETE
  USING (public.app_is_admin());
CREATE TRIGGER catalog_guard BEFORE UPDATE ON public.foods
  FOR EACH ROW EXECUTE FUNCTION public.catalog_guard();

-- migrate:down
-- Lossy once several users have data: the global uniques (day date, recipe name, client IDs) come
-- back, so the down fails if two users collide. Dev and the D1 rehearsal only.
DROP TRIGGER catalog_guard ON public.foods;
DROP POLICY catalog_delete ON public.foods;
DROP POLICY catalog_edit ON public.foods;
DROP POLICY catalog_add ON public.foods;
DROP POLICY catalog_read ON public.foods;
ALTER TABLE public.foods DISABLE ROW LEVEL SECURITY;

DROP POLICY owner ON public.composite_food_items;
ALTER TABLE public.composite_food_items DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.saved_meal_items;
ALTER TABLE public.saved_meal_items DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.meal_items;
ALTER TABLE public.meal_items DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.day_logs;
ALTER TABLE public.day_logs DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.planned_meals;
ALTER TABLE public.planned_meals DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.composite_foods;
ALTER TABLE public.composite_foods DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.saved_meals;
ALTER TABLE public.saved_meals DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.meals;
ALTER TABLE public.meals DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.days;
ALTER TABLE public.days DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.plans;
ALTER TABLE public.plans DISABLE ROW LEVEL SECURITY;

DROP INDEX public.uq_planned_meals_user_client_id;
CREATE UNIQUE INDEX uq_planned_meals_client_id ON public.planned_meals USING btree (client_id) WHERE client_id IS NOT NULL;
DROP INDEX public.uq_meals_user_client_id;
CREATE UNIQUE INDEX uq_meals_client_id ON public.meals USING btree (client_id) WHERE client_id IS NOT NULL;
DROP INDEX public.uq_composite_foods_user_name;
CREATE UNIQUE INDEX idx_composite_foods_name ON public.composite_foods USING btree (name);
DROP INDEX public.uq_days_user_date;
CREATE UNIQUE INDEX idx_days_date ON public.days USING btree (date);

ALTER TABLE public.saved_meal_items DROP CONSTRAINT fk_saved_meal_items_composite_food;
ALTER TABLE public.meal_items DROP CONSTRAINT fk_meal_items_composite_food;
ALTER TABLE public.planned_meals DROP CONSTRAINT fk_planned_meals_saved_meal;

DROP INDEX public.idx_composite_food_items_user_composite;
ALTER TABLE public.composite_food_items DROP CONSTRAINT fk_composite_foods_items;
ALTER TABLE public.composite_food_items ADD CONSTRAINT fk_composite_foods_items FOREIGN KEY (composite_food_id)
  REFERENCES public.composite_foods (id) ON DELETE CASCADE;
DROP INDEX public.idx_saved_meal_items_user_saved_meal;
ALTER TABLE public.saved_meal_items DROP CONSTRAINT fk_saved_meals_items;
ALTER TABLE public.saved_meal_items ADD CONSTRAINT fk_saved_meals_items FOREIGN KEY (saved_meal_id)
  REFERENCES public.saved_meals (id) ON DELETE CASCADE;
DROP INDEX public.idx_meal_items_user_meal;
ALTER TABLE public.meal_items DROP CONSTRAINT fk_meals_items;
ALTER TABLE public.meal_items ADD CONSTRAINT fk_meals_items FOREIGN KEY (meal_id)
  REFERENCES public.meals (id) ON DELETE CASCADE;
DROP INDEX public.idx_day_logs_user_day;
ALTER TABLE public.day_logs DROP CONSTRAINT fk_day_logs_meal;
ALTER TABLE public.day_logs ADD CONSTRAINT fk_day_logs_meal FOREIGN KEY (meal_id) REFERENCES public.meals (id);
ALTER TABLE public.day_logs DROP CONSTRAINT fk_days_logs;
ALTER TABLE public.day_logs ADD CONSTRAINT fk_days_logs FOREIGN KEY (day_id) REFERENCES public.days (id);
DROP INDEX public.idx_planned_meals_user_day;
ALTER TABLE public.planned_meals DROP CONSTRAINT fk_planned_meals_meal;
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_planned_meals_meal FOREIGN KEY (meal_id) REFERENCES public.meals (id);
ALTER TABLE public.planned_meals DROP CONSTRAINT fk_days_planned_meals;
ALTER TABLE public.planned_meals ADD CONSTRAINT fk_days_planned_meals FOREIGN KEY (day_id) REFERENCES public.days (id);
ALTER TABLE public.days DROP CONSTRAINT fk_days_plan;
ALTER TABLE public.days ADD CONSTRAINT fk_days_plan FOREIGN KEY (plan_id) REFERENCES public.plans (id);

ALTER TABLE public.composite_foods DROP CONSTRAINT uq_composite_foods_id_user;
ALTER TABLE public.saved_meals DROP CONSTRAINT uq_saved_meals_id_user;
ALTER TABLE public.meals DROP CONSTRAINT uq_meals_id_user;
ALTER TABLE public.days DROP CONSTRAINT uq_days_id_user;
ALTER TABLE public.plans DROP CONSTRAINT uq_plans_id_user;

ALTER TABLE public.composite_food_items DROP CONSTRAINT fk_composite_food_items_user;
ALTER TABLE public.composite_food_items DROP COLUMN user_id;
ALTER TABLE public.saved_meal_items DROP CONSTRAINT fk_saved_meal_items_user;
ALTER TABLE public.saved_meal_items DROP COLUMN user_id;
ALTER TABLE public.meal_items DROP CONSTRAINT fk_meal_items_user;
ALTER TABLE public.meal_items DROP COLUMN user_id;
ALTER TABLE public.day_logs DROP CONSTRAINT fk_day_logs_user;
ALTER TABLE public.day_logs DROP COLUMN user_id;
ALTER TABLE public.planned_meals DROP CONSTRAINT fk_planned_meals_user;
ALTER TABLE public.planned_meals DROP COLUMN user_id;
DROP INDEX public.idx_composite_foods_user_id;
ALTER TABLE public.composite_foods DROP CONSTRAINT fk_composite_foods_user;
ALTER TABLE public.composite_foods DROP COLUMN user_id;
DROP INDEX public.idx_saved_meals_user_id;
ALTER TABLE public.saved_meals DROP CONSTRAINT fk_saved_meals_user;
ALTER TABLE public.saved_meals DROP COLUMN user_id;
DROP INDEX public.idx_meals_user_id;
ALTER TABLE public.meals DROP CONSTRAINT fk_meals_user;
ALTER TABLE public.meals DROP COLUMN user_id;
DROP INDEX public.idx_days_user_id;
ALTER TABLE public.days DROP CONSTRAINT fk_days_user;
ALTER TABLE public.days DROP COLUMN user_id;
DROP INDEX public.idx_plans_user_id;
ALTER TABLE public.plans DROP CONSTRAINT fk_plans_user;
ALTER TABLE public.plans DROP COLUMN user_id;
DROP INDEX public.idx_foods_created_by;
ALTER TABLE public.foods DROP CONSTRAINT fk_foods_created_by;
ALTER TABLE public.foods DROP COLUMN created_by;
