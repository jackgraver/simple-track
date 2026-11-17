// Package migrations holds the SQL migrations, embedded into the server
// (startup schema check) and cmd/migrate (which applies them). See
// internal/migrate for the file format; create new ones with
// `migrate new <name>`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
