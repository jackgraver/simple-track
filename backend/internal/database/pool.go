package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Pool is the database/sql connection pool configuration. The defaults are sized for one
// server process on a small VPS serving a handful of family users, sharing Postgres'
// default max_connections (100) with deploy-time tools (migrate) and a psql session or two.
// Reasoning in docs/multi-user/DECISIONS.md ("Connection pool").
type Pool struct {
	MaxOpen     int           // DB_MAX_OPEN_CONNS: hard cap; requests beyond it wait for a free connection
	MaxIdle     int           // DB_MAX_IDLE_CONNS: kept open between bursts (capped at MaxOpen)
	MaxLifetime time.Duration // DB_CONN_MAX_LIFETIME: recycle connections (server-side memory, config changes)
	MaxIdleTime time.Duration // DB_CONN_MAX_IDLE_TIME: close idle connections after this long
}

// DefaultPool is used for any DB_* variable that isn't set.
var DefaultPool = Pool{
	MaxOpen:     20,
	MaxIdle:     10,
	MaxLifetime: 30 * time.Minute,
	MaxIdleTime: 5 * time.Minute,
}

// PoolFromEnv reads DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS (integers) and DB_CONN_MAX_LIFETIME,
// DB_CONN_MAX_IDLE_TIME (Go durations, e.g. 30m). Unset or empty means the default; an
// invalid value is an error, so a typo can't silently fall back to something else.
func PoolFromEnv() (Pool, error) {
	p := DefaultPool
	var err error
	if p.MaxOpen, err = intEnv("DB_MAX_OPEN_CONNS", p.MaxOpen, 1); err != nil {
		return p, err
	}
	if p.MaxIdle, err = intEnv("DB_MAX_IDLE_CONNS", p.MaxIdle, 0); err != nil {
		return p, err
	}
	if p.MaxLifetime, err = durationEnv("DB_CONN_MAX_LIFETIME", p.MaxLifetime); err != nil {
		return p, err
	}
	if p.MaxIdleTime, err = durationEnv("DB_CONN_MAX_IDLE_TIME", p.MaxIdleTime); err != nil {
		return p, err
	}
	if p.MaxIdle > p.MaxOpen {
		p.MaxIdle = p.MaxOpen
	}
	return p, nil
}

func (p Pool) apply(db *sql.DB) {
	db.SetMaxOpenConns(p.MaxOpen)
	db.SetMaxIdleConns(p.MaxIdle)
	db.SetConnMaxLifetime(p.MaxLifetime)
	db.SetConnMaxIdleTime(p.MaxIdleTime)
}

func (p Pool) String() string {
	return fmt.Sprintf("pool: max open %d, max idle %d, lifetime %s, idle time %s",
		p.MaxOpen, p.MaxIdle, p.MaxLifetime, p.MaxIdleTime)
}

func intEnv(key string, fallback, min int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < min {
		return 0, fmt.Errorf("%s: want an integer >= %d, got %q", key, min, v)
	}
	return n, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s: want a positive duration such as 30m, got %q", key, v)
	}
	return d, nil
}
