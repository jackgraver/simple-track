package database

import (
	"context"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// AppRole is the login role the server connects as. Migration 0010_app_role creates it (NOLOGIN,
// no password) with its grants and timeouts; EnsureAppLogin, run by `migrate up`, gives it LOGIN
// and the password from the app DSN. The password never appears in a migration.
const AppRole = "simpletrack_app"

// CheckAppRole reports why the connection's current role must not run the server: superuser,
// BYPASSRLS, owner of a table in the public schema (directly or by inheriting the owner's
// role: RLS treats both as the owner), or a member of a superuser/BYPASSRLS role (one SET ROLE
// away). Any of these switches off row-level security. Nil means the role is fine.
func CheckAppRole(ctx context.Context, db rowQuerier) error {
	var (
		user            string
		super, bypass   bool
		owned, powerful int
	)
	err := db.QueryRowContext(ctx, `
		SELECT current_user, r.rolsuper, r.rolbypassrls,
		       (SELECT count(*) FROM pg_tables WHERE schemaname = 'public'
		          AND pg_has_role(current_user, tableowner, 'USAGE')),
		       (SELECT count(*) FROM pg_roles p WHERE p.rolname <> current_user
		          AND (p.rolsuper OR p.rolbypassrls) AND pg_has_role(current_user, p.oid, 'MEMBER'))
		FROM pg_roles r WHERE r.rolname = current_user`).Scan(&user, &super, &bypass, &owned, &powerful)
	if err != nil {
		return fmt.Errorf("read current role: %w", err)
	}
	var why []string
	if super {
		why = append(why, "is a superuser")
	}
	if bypass {
		why = append(why, "has BYPASSRLS")
	}
	if owned > 0 {
		why = append(why, fmt.Sprintf("owns (or inherits the owner of) %d table(s)", owned))
	}
	if powerful > 0 {
		why = append(why, fmt.Sprintf("is a member of %d superuser/BYPASSRLS role(s)", powerful))
	}
	if len(why) == 0 {
		return nil
	}
	return fmt.Errorf("database role %q %s, so row-level security doesn't apply to it; the server must connect as %s (DATABASE_URL / DATABASE_URL_PRODUCTION) and only the tools as the owner (DATABASE_OWNER_URL)",
		user, strings.Join(why, ", "), AppRole)
}

// rowQuerier is satisfied by *sql.DB, *sql.Conn and *sql.Tx.
type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// EnsureAppLogin makes appDSN able to log in, using owner (a connection as the schema owner or
// a role with CREATEROLE). It does nothing when appDSN's user isn't AppRole (e.g. a single-role
// setup) or when appDSN can already connect. Otherwise it sets LOGIN and appDSN's password on
// the role, sent as a SCRAM-SHA-256 verifier so the plain password never reaches Postgres
// (or its statement log). The role itself must already exist (migration 0010).
//
// Checking first means a normal deploy changes nothing, and parallel runs against one cluster
// (e2e) don't race on the shared role row. Changing the password in the environment and running
// `migrate up` rotates it.
func EnsureAppLogin(ctx context.Context, owner *sql.DB, appDSN string) (changed bool, err error) {
	return ensureLogin(ctx, owner, appDSN, AppRole)
}

func ensureLogin(ctx context.Context, owner *sql.DB, dsn, role string) (bool, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return false, fmt.Errorf("parse app DSN: %w", err)
	}
	if cfg.User != role {
		return false, nil
	}
	if cfg.Password == "" {
		return false, fmt.Errorf("the app DSN for %s has no password", role)
	}
	connErr := tryConnect(ctx, cfg)
	if connErr == nil {
		return false, nil
	}
	if err := needsBootstrap(ctx, owner, role, connErr); err != nil {
		return false, err
	}
	verifier, err := scramVerifier(cfg.Password)
	if err != nil {
		return false, err
	}
	// ALTER ROLE takes no bind parameters; both values are quoted by Postgres' own functions.
	var stmt string
	if err := owner.QueryRowContext(ctx, `SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', $1::text, $2::text)`,
		role, verifier).Scan(&stmt); err != nil {
		return false, err
	}
	if _, err := owner.ExecContext(ctx, stmt); err != nil {
		// Another `migrate up` against the same cluster (e2e templates built in parallel) may have
		// updated the role at the same moment ("tuple concurrently updated"). If the DSN works
		// now, that run did our job.
		if tryConnect(ctx, cfg) == nil {
			return false, nil
		}
		return false, fmt.Errorf("set %s login: %w", role, err)
	}
	if err := tryConnect(ctx, cfg); err != nil {
		return true, fmt.Errorf("connect as %s after setting its password: %w", role, err)
	}
	return true, nil
}

// needsBootstrap decides whether a failed login is ours to fix. Wrong password (28P01): yes, the
// DSN's password is the source of truth (rotation). Can't log in (28000): only while the role is
// still as migration 0010 left it, NOLOGIN with no password. A role with a password that is
// NOLOGIN was switched off by someone on purpose, and a 28000 for a LOGIN role is pg_hba or
// similar; neither is fixed by resetting the password.
func needsBootstrap(ctx context.Context, owner *sql.DB, role string, connErr error) error {
	var pgErr *pgconn.PgError
	if !errors.As(connErr, &pgErr) || (pgErr.Code != "28P01" && pgErr.Code != "28000") {
		return fmt.Errorf("connect as %s: %w", role, connErr)
	}
	if pgErr.Code == "28P01" {
		return nil
	}
	var canLogin bool
	if err := owner.QueryRowContext(ctx, `SELECT rolcanlogin FROM pg_roles WHERE rolname = $1`, role).Scan(&canLogin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role %s doesn't exist: run every migration first (0010 creates it)", role)
		}
		return err
	}
	if canLogin {
		return fmt.Errorf("connect as %s (it can log in, so this isn't a missing password; check pg_hba.conf): %w", role, connErr)
	}
	// pg_authid is readable by superusers only; without it, assume the 0010 state.
	var hasPassword bool
	if err := owner.QueryRowContext(ctx, `SELECT rolpassword IS NOT NULL FROM pg_authid WHERE rolname = $1`, role).Scan(&hasPassword); err == nil && hasPassword {
		return fmt.Errorf("role %s has a password but NOLOGIN, so it was disabled deliberately; re-enable it by hand (ALTER ROLE %s LOGIN) if that's intended", role, role)
	}
	return nil
}

func tryConnect(ctx context.Context, cfg *pgx.ConnConfig) error {
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	return conn.Close(ctx)
}

// scramVerifier builds the SCRAM-SHA-256 secret Postgres stores for a password (RFC 5802 /
// RFC 7677, the format of pg_authid.rolpassword). Postgres SASLpreps the password first; for
// the ASCII passwords we generate that is the identity, so non-ASCII passwords are refused.
func scramVerifier(password string) (string, error) {
	for _, r := range password {
		if r < 0x21 || r > 0x7e {
			return "", errors.New("the app role password must be printable ASCII without spaces (use e.g. `openssl rand -hex 24`)")
		}
	}
	const iterations = 4096 // Postgres' default (scram_iterations)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return scramVerifierWithSalt(password, salt, iterations)
}

func scramVerifierWithSalt(password string, salt []byte, iterations int) (string, error) {
	salted, err := pbkdf2.Key(sha256.New, password, salt, iterations, sha256.Size)
	if err != nil {
		return "", err
	}
	mac := func(key []byte, msg string) []byte {
		h := hmac.New(sha256.New, key)
		h.Write([]byte(msg))
		return h.Sum(nil)
	}
	clientKey := mac(salted, "Client Key")
	storedKey := sha256.Sum256(clientKey)
	serverKey := mac(salted, "Server Key")
	b64 := base64.StdEncoding.EncodeToString
	return fmt.Sprintf("SCRAM-SHA-256$%d:%s$%s:%s", iterations, b64(salt), b64(storedKey[:]), b64(serverKey)), nil
}
