-- Accounts for multi-user (lane B1).
-- is_admin: the owner. Set here for the one existing account; with none (fresh or e2e DB) nobody
-- is admin until devuser/e2eseed/an admin sets it. More than one live account means the owner
-- can't be told apart, so the migration stops and asks for it to be set by hand.
-- token_version: copied into every JWT; bumping it revokes the user's sessions. Starts at 1 so
-- a token without a version never matches.
-- disabled_at: set means no sign-in and every token rejected.

-- migrate:up
ALTER TABLE public.users ADD COLUMN is_admin boolean NOT NULL DEFAULT false;
ALTER TABLE public.users ADD COLUMN token_version bigint NOT NULL DEFAULT 1;
ALTER TABLE public.users ADD COLUMN disabled_at timestamp with time zone;

DO $$
DECLARE live bigint;
BEGIN
  SELECT count(*) INTO live FROM public.users WHERE deleted_at IS NULL;
  IF live > 1 THEN
    RAISE EXCEPTION 'users has % live accounts, expected at most one (the owner): resolve by hand before migrating (docs/multi-user/DECISIONS.md)', live;
  END IF;
  UPDATE public.users SET is_admin = true WHERE deleted_at IS NULL;
END $$;

-- migrate:down
ALTER TABLE public.users DROP COLUMN disabled_at;
ALTER TABLE public.users DROP COLUMN token_version;
ALTER TABLE public.users DROP COLUMN is_admin;
