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
// BYPASSRLS, or owner of a table in the public schema. Any of these silently switches off
// row-level security. Nil means the role is fine.
func CheckAppRole(ctx context.Context, db rowQuerier) error {
	var (
		user          string
		super, bypass bool
		owned         int
	)
	err := db.QueryRowContext(ctx, `
		SELECT current_user, r.rolsuper, r.rolbypassrls,
		       (SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tableowner = current_user)
		FROM pg_roles r WHERE r.rolname = current_user`).Scan(&user, &super, &bypass, &owned)
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
		why = append(why, fmt.Sprintf("owns %d table(s)", owned))
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
	if err := tryConnect(ctx, cfg); err == nil {
		return false, nil
	} else if !isAuthError(err) {
		return false, fmt.Errorf("connect as %s: %w", role, err)
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
		return false, fmt.Errorf("set %s login: %w", role, err)
	}
	if err := tryConnect(ctx, cfg); err != nil {
		return true, fmt.Errorf("connect as %s after setting its password: %w", role, err)
	}
	return true, nil
}

func tryConnect(ctx context.Context, cfg *pgx.ConnConfig) error {
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	return conn.Close(ctx)
}

// isAuthError: wrong password (28P01) or role can't log in / doesn't exist (28000).
func isAuthError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "28P01" || pgErr.Code == "28000")
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
