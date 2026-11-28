-- The server's login role, simpletrack_app (docs/multi-user/DECISIONS.md, "App DB role").
-- DML on every table, no ownership, no BYPASSRLS, so row-level security will apply to it.
-- This migration runs as the schema owner, and ALTER DEFAULT PRIVILEGES below covers every
-- table and sequence that role creates later, so future migrations need no GRANTs.
--
-- No password here: the role is created NOLOGIN. `migrate up` (database.EnsureAppLogin) gives
-- it LOGIN and the password from the app DSN in the environment.
--
-- Roles and role settings are cluster-wide, but migrations run per database (tests and the e2e
-- suite migrate many databases on one server, sometimes at once). So the cluster-wide steps only
-- act when something is missing, and tolerate a concurrent migration in another database doing
-- the same thing first.

-- migrate:up
DO $$
DECLARE
  r record;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'simpletrack_app') THEN
    BEGIN
      CREATE ROLE simpletrack_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
    EXCEPTION WHEN duplicate_object OR unique_violation THEN
      NULL; -- created by a concurrent migration in another database
    END;
  END IF;

  -- A pre-existing role with powers it shouldn't have is stripped of them.
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'simpletrack_app'
             AND (rolsuper OR rolbypassrls OR rolcreatedb OR rolcreaterole OR rolreplication)) THEN
    ALTER ROLE simpletrack_app NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
  END IF;

  -- ...and of any role memberships: inheriting the table owner's role would make RLS treat the
  -- app as the owner. (The server's CheckAppRole refuses that too.)
  FOR r IN SELECT g.rolname FROM pg_auth_members m
           JOIN pg_roles g ON g.oid = m.roleid
           WHERE m.member = 'simpletrack_app'::regrole LOOP
    EXECUTE format('REVOKE %I FROM simpletrack_app', r.rolname);
  END LOOP;

  -- Timeouts apply to every session the role opens (not to the owner, so migrations are
  -- unaffected). statement_timeout caps a runaway query; idle_in_transaction_session_timeout
  -- ends a session that left a transaction open (the per-request transaction makes that
  -- possible), releasing its locks. Values: DECISIONS.md, "Connection pool".
  IF NOT EXISTS (SELECT 1 FROM pg_db_role_setting s
                 WHERE s.setrole = 'simpletrack_app'::regrole AND s.setdatabase = 0
                   AND s.setconfig @> ARRAY['statement_timeout=15s']) THEN
    BEGIN
      ALTER ROLE simpletrack_app SET statement_timeout = '15s';
    EXCEPTION WHEN unique_violation OR internal_error THEN
      NULL; -- "tuple concurrently updated": a concurrent migration set it
    END;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_db_role_setting s
                 WHERE s.setrole = 'simpletrack_app'::regrole AND s.setdatabase = 0
                   AND s.setconfig @> ARRAY['idle_in_transaction_session_timeout=30s']) THEN
    BEGIN
      ALTER ROLE simpletrack_app SET idle_in_transaction_session_timeout = '30s';
    EXCEPTION WHEN unique_violation OR internal_error THEN
      NULL;
    END;
  END IF;
END $$;

-- Per database from here on.
REVOKE CREATE ON SCHEMA public FROM PUBLIC; -- already the default since Postgres 15
GRANT USAGE ON SCHEMA public TO simpletrack_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO simpletrack_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO simpletrack_app;
-- The server only reads migration history (startup schema check).
REVOKE INSERT, UPDATE, DELETE ON public.schema_migrations FROM simpletrack_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO simpletrack_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO simpletrack_app;

-- migrate:down
-- Leaves REVOKE CREATE ON SCHEMA public FROM PUBLIC in place: it is the Postgres 15+ default
-- and nothing should depend on PUBLIC creating tables.
-- Leaves the role itself (and its cluster-wide settings) in place: other databases on the same
-- server may use it. Drop it by hand once no database grants it anything.
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM simpletrack_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM simpletrack_app;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM simpletrack_app;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM simpletrack_app;
REVOKE USAGE ON SCHEMA public FROM simpletrack_app;
