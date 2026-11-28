-- Accounts and setup links (lane C6, docs/multi-user/DECISIONS.md "C6").
--
-- users: usernames are unique case-insensitively (BACKLOG #14). New usernames are stored in
-- lowercase by the app; existing ones are left as they are, so the owner's login keeps working
-- (login compares lower(username)). An empty email no longer counts as a taken email, so
-- accounts created with a username only don't collide.
--
-- user_setup_tokens: one-time links to set or reset a password. Only a SHA-256 of the token is
-- stored. RLS: the admin sees and writes every row; anyone else sees only the row whose hash
-- the request put in app.setup_token_hash, i.e. a caller holding the token itself.
--
-- auth_events: append-only audit log. RLS: anyone may insert (failed logins have no identity),
-- only the admin reads, nobody updates or deletes (no policy, and the privileges are revoked).
-- The app inserts without RETURNING, since RETURNING would need the SELECT policy to pass.

-- migrate:up
DROP INDEX public.idx_users_username;
CREATE UNIQUE INDEX idx_users_username_lower ON public.users (lower(username));

UPDATE public.users SET email = NULL WHERE email = '';
DROP INDEX public.idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON public.users (email) WHERE email IS NOT NULL AND email <> '';

CREATE TABLE public.user_setup_tokens (
  id          bigserial PRIMARY KEY,
  user_id     bigint NOT NULL REFERENCES public.users (id) ON DELETE CASCADE,
  token_hash  text NOT NULL UNIQUE,
  purpose     text NOT NULL CHECK (purpose IN ('setup', 'reset')),
  created_by  bigint REFERENCES public.users (id) ON DELETE SET NULL,
  created_at  timestamp with time zone NOT NULL DEFAULT now(),
  expires_at  timestamp with time zone NOT NULL,
  used_at     timestamp with time zone,
  revoked_at  timestamp with time zone
);
CREATE INDEX idx_user_setup_tokens_user ON public.user_setup_tokens (user_id);

ALTER TABLE public.user_setup_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY admin_all ON public.user_setup_tokens
  USING      (current_setting('app.is_admin', true) = 'true')
  WITH CHECK (current_setting('app.is_admin', true) = 'true');
CREATE POLICY token_holder ON public.user_setup_tokens
  USING      (token_hash = current_setting('app.setup_token_hash', true))
  WITH CHECK (token_hash = current_setting('app.setup_token_hash', true));

CREATE TABLE public.auth_events (
  id          bigserial PRIMARY KEY,
  created_at  timestamp with time zone NOT NULL DEFAULT now(),
  event       text NOT NULL,
  user_id     bigint REFERENCES public.users (id) ON DELETE SET NULL,
  actor_id    bigint REFERENCES public.users (id) ON DELETE SET NULL,
  client_ip   text NOT NULL DEFAULT '',
  detail      text NOT NULL DEFAULT ''
);
CREATE INDEX idx_auth_events_created ON public.auth_events (created_at DESC, id DESC);
CREATE INDEX idx_auth_events_user ON public.auth_events (user_id);

ALTER TABLE public.auth_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY append ON public.auth_events FOR INSERT WITH CHECK (true);
CREATE POLICY admin_read ON public.auth_events FOR SELECT
  USING (current_setting('app.is_admin', true) = 'true');
REVOKE UPDATE, DELETE, TRUNCATE ON public.auth_events FROM simpletrack_app;

-- migrate:down
DROP TABLE public.auth_events;
DROP TABLE public.user_setup_tokens;
DROP INDEX public.idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);
DROP INDEX public.idx_users_username_lower;
CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);
