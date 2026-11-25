//go:build e2e

package e2e

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// fault makes DB statements on Table fail. The error is returned by gorm before the statement runs
// (inside its transaction, so it rolls back), and the real repository/service/handler code turns it
// into whatever response it really produces.
type fault struct {
	ID    string `json:"-"`
	Table string `json:"table" binding:"required"` // e.g. "logged_sets"
	Op    string `json:"op" binding:"required"`    // create | query | update | delete | row | raw
	// Code, if set, returns a *pgconn.PgError with this SQLSTATE (e.g. 23505 unique_violation,
	// 40001 serialization_failure, 57014 query_canceled) instead of a plain error.
	Code    string `json:"code"`
	Message string `json:"message"`
	Times   int    `json:"times"` // 0 = every matching statement; n = the next n
	Hits    int    `json:"-"`
}

var faultOps = []string{"create", "query", "update", "delete", "row", "raw"}

func (s *state) addFault(f fault) (string, error) {
	if !slices.Contains(faultOps, f.Op) {
		return "", fmt.Errorf("op %q: want one of %s", f.Op, strings.Join(faultOps, ", "))
	}
	if !s.db.Migrator().HasTable(f.Table) {
		return "", fmt.Errorf("no table %q", f.Table)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f.ID = s.id()
	s.faults = append(s.faults, &f)
	return f.ID, nil
}

func (f *fault) err() error {
	msg := f.Message
	if msg == "" {
		msg = fmt.Sprintf("e2e fault: %s %s", f.Op, f.Table)
	}
	if f.Code != "" {
		return &pgconn.PgError{Severity: "ERROR", Code: f.Code, Message: msg, TableName: f.Table}
	}
	return errors.New(msg)
}

// match returns the error for the first live fault on (table, op), counting the hit.
func (s *state) match(table, op, sql string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.faults {
		if f.Op != op || (f.Times > 0 && f.Hits >= f.Times) {
			continue
		}
		// Raw/row statements have no parsed table; match on the SQL text instead.
		if f.Table == table || (table == "" && sql != "" && strings.Contains(sql, f.Table)) {
			f.Hits++
			return f.err()
		}
	}
	return nil
}

func (s *state) registerFaultHooks(db *gorm.DB) error {
	cb := db.Callback()
	hook := func(op string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			if tx.Error != nil {
				return
			}
			if err := s.match(tx.Statement.Table, op, tx.Statement.SQL.String()); err != nil {
				_ = tx.AddError(err)
			}
		}
	}
	return errors.Join(
		cb.Create().Before("gorm:create").Register("e2e:fault", hook("create")),
		cb.Query().Before("gorm:query").Register("e2e:fault", hook("query")),
		cb.Update().Before("gorm:update").Register("e2e:fault", hook("update")),
		cb.Delete().Before("gorm:delete").Register("e2e:fault", hook("delete")),
		cb.Row().Before("gorm:row").Register("e2e:fault", hook("row")),
		cb.Raw().Before("gorm:raw").Register("e2e:fault", hook("raw")),
	)
}
