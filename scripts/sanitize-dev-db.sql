BEGIN;

DO $$
BEGIN
    IF to_regclass('public.users') IS NOT NULL THEN
        DELETE FROM users;
    END IF;
END
$$;

COMMIT;
