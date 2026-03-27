# Internal application packages

Application code lives under `internal/` so it is not importable by other Go modules.

Run the API from the backend module root:

```bash
go run ./cmd/server
```

Seed workout demo data (destructive — drops workout tables):

```bash
go run ./cmd/seed
```

## NUMBER 6 HERE

Optional follow-ups (your choice): introduce service interfaces for testing, table-driven tests, `httptest` for HTTP handlers, or other DI patterns.
