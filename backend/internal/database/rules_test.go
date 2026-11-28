package database

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Owner-kit rules for request code (BACKLOG #1 and #2; DECISIONS.md "B3 owner kit"). The
// allowlists hold today's sites as a count per file. A file may go below its count (lanes shrink
// them as they convert), never above, and an entry for a file that no longer has any site, or no
// longer exists, is fine. When you fix sites, lower or delete the entry in the same PR.

// gorm Create/Save/FirstOrCreate/CreateInBatches without Omit(clause.Associations) in the same
// chain: gorm would also write whatever nested structs the value carries (a food inside a saved
// meal item, BACKLOG #1). Sites that predate the rule; each module's C lane converts its own.
var createWithAssociationsAllowed = map[string]int{
	"internal/core/auth/repository/user_repository.go": 1, // C6 (users: no associations)
}

// Advisory locks taken without LockKey (BACKLOG #2): C2 and C3 move them to database.XactLock.
var rawAdvisoryLockAllowed = map[string]int{}

// Code that sets user_id / created_by (or a field named UserID / CreatedBy) other than the
// Owner / Creator create hooks. Each entry needs a reason.
var ownerFieldWriteAllowed = map[string]int{
	// database.Identity{UserID: ...}: the request identity itself, not a row's owner.
	"internal/core/auth/auth_middleware.go": 1,
	// C6: users, user_setup_tokens and auth_events are identity tables, not owned data. Their
	// user_id names the account a link or event is about (often not the caller: an admin acts on
	// someone else, a failed login has no caller), so the Owner hook can't stamp it. RLS guards
	// them instead (DECISIONS, C6).
	"internal/core/auth/auth_api.go":                 1,
	"internal/core/auth/controller/account_route.go": 1,
	"internal/core/auth/controller/auth_route.go":    2,
	"internal/core/auth/services/account_service.go": 10,
	"internal/core/auth/services/auth_service.go":    3,
}

var (
	gormWriteMethods = map[string]bool{"Create": true, "CreateInBatches": true, "Save": true, "FirstOrCreate": true}
	bindMethods      = regexp.MustCompile(`^(ShouldBind|Bind|MustBind)\w*$`)
	ownerFieldWrite  = regexp.MustCompile(`\.(UserID|CreatedBy)\s*=[^=]|\b(UserID|CreatedBy)\s*:[^=]|"(user_id|created_by)"\s*[:,]`)
	rawAdvisoryLock  = regexp.MustCompile(`pg_(try_)?advisory`)
	withIdentityCall = regexp.MustCompile(`database\.WithIdentity\(`)
)

type goFile struct {
	rel  string
	src  []byte
	file *ast.File
	fset *token.FileSet
}

// requestCode parses the non-test Go files under internal/core and internal/utils (the same
// scope as TestRequestCodeUsesFrom), testutil packages excluded.
func requestCode(t *testing.T, root string) []goFile {
	t.Helper()
	var files []goFile
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		inScope := false
		for _, p := range requestPaths {
			inScope = inScope || strings.HasPrefix(rel, p)
		}
		if !inScope || strings.Contains(rel, "/testutil/") || strings.Contains(rel, "/testdata/") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}
		files = append(files, goFile{rel: rel, src: src, file: f, fset: fset})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestOwnerKitRules(t *testing.T) {
	files := requestCode(t, filepath.Join("..", ".."))
	if len(files) < 50 {
		t.Fatalf("only %d files checked; did the layout move?", len(files))
	}
	r := checkOwnerKitRules(files)
	for _, e := range r.errors {
		t.Error(e)
	}
	checkCounts(t, "create/save without Omit(clause.Associations) (BACKLOG #1)", r.createWithAssoc, createWithAssociationsAllowed)
	checkCounts(t, "pg_advisory lock without database.XactLock/LockKey (BACKLOG #2)", r.rawLocks, rawAdvisoryLockAllowed)
	checkCounts(t, "UserID/CreatedBy/user_id/created_by written outside the Owner/Creator hooks (BACKLOG #1)", r.ownerWrites, ownerFieldWriteAllowed)
}

type ruleResult struct {
	errors                                 []string
	createWithAssoc, rawLocks, ownerWrites map[string][]string // file -> sites
}

// checkCounts fails for a file with more sites than its allowlist entry.
func checkCounts(t *testing.T, rule string, sites map[string][]string, allowed map[string]int) {
	t.Helper()
	files := make([]string, 0, len(sites))
	for f := range sites {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		if got, max := len(sites[f]), allowed[f]; got > max {
			t.Errorf("%s: %d new site(s), allowlist has %d for this file:\n\t%s",
				rule, got-max, max, strings.Join(sites[f], "\n\t"))
		}
	}
}

func checkOwnerKitRules(files []goFile) ruleResult {
	r := ruleResult{
		createWithAssoc: map[string][]string{},
		rawLocks:        map[string][]string{},
		ownerWrites:     map[string][]string{},
	}
	owned := ownerBearingTypes(files)
	for _, gf := range files {
		pos := func(p token.Pos) string { return fmt.Sprintf("%s:%d", gf.rel, gf.fset.Position(p).Line) }
		for _, loc := range ownerFieldWrite.FindAllIndex(gf.src, -1) {
			line := 1 + strings.Count(string(gf.src[:loc[0]]), "\n")
			r.ownerWrites[gf.rel] = append(r.ownerWrites[gf.rel], fmt.Sprintf("%s:%d", gf.rel, line))
		}
		for _, loc := range rawAdvisoryLock.FindAllIndex(gf.src, -1) {
			line := 1 + strings.Count(string(gf.src[:loc[0]]), "\n")
			r.rawLocks[gf.rel] = append(r.rawLocks[gf.rel], fmt.Sprintf("%s:%d", gf.rel, line))
		}
		for _, loc := range withIdentityCall.FindAllIndex(gf.src, -1) {
			line := 1 + strings.Count(string(gf.src[:loc[0]]), "\n")
			r.errors = append(r.errors, fmt.Sprintf("%s:%d: database.WithIdentity is for tests and tools; request code gets its user from AuthMiddleware", gf.rel, line))
		}
		imports := map[string]bool{}
		for _, im := range gf.file.Imports {
			p := strings.Trim(im.Path.Value, `"`)
			name := p[strings.LastIndex(p, "/")+1:]
			if im.Name != nil {
				name = im.Name.Name
			}
			imports[name] = true
		}

		ast.Inspect(gf.file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.TypeSpec:
				checkBothEmbedded(n, pos, &r)
			case *ast.FuncDecl:
				checkBeforeCreate(n, owned, pos, &r)
				checkBinds(n, owned, pos, &r)
			case *ast.CallExpr:
				sel, ok := n.Fun.(*ast.SelectorExpr)
				if !ok || !gormWriteMethods[sel.Sel.Name] {
					return true
				}
				if id, ok := sel.X.(*ast.Ident); ok && imports[id.Name] {
					return true // os.Create and friends
				}
				if !chainOmitsAssociations(sel.X) {
					r.createWithAssoc[gf.rel] = append(r.createWithAssoc[gf.rel], pos(n.Pos())+": "+sel.Sel.Name)
				}
			}
			return true
		})
	}
	return r
}

// chainOmitsAssociations reports whether a gorm call chain contains Omit(clause.Associations).
func chainOmitsAssociations(x ast.Expr) bool {
	for {
		switch e := x.(type) {
		case *ast.CallExpr:
			if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Omit" {
					for _, a := range e.Args {
						if s, ok := a.(*ast.SelectorExpr); ok && s.Sel.Name == "Associations" {
							return true
						}
					}
				}
				x = sel.X
				continue
			}
			return false
		case *ast.SelectorExpr:
			x = e.X
		case *ast.ParenExpr:
			x = e.X
		default:
			return false
		}
	}
}

// typeInfo is what the rules know about a named struct type: whether it has an owner field
// itself, and the names of the types its fields use.
type typeInfo struct {
	direct bool
	uses   []string
}

// ownerBearingTypes returns the names of struct types that carry an owner: a field named UserID
// or CreatedBy, an embedded (database.)Owner / Creator, or a field (also slice, pointer, map) of
// such a type. Types are matched by name only, across packages, so a name clash errs on the side
// of flagging.
func ownerBearingTypes(files []goFile) map[string]bool {
	info := map[string]*typeInfo{}
	for _, gf := range files {
		for _, d := range gf.file.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts := s.(*ast.TypeSpec)
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				ti := info[ts.Name.Name]
				if ti == nil {
					ti = &typeInfo{}
					info[ts.Name.Name] = ti
				}
				collectStruct(st, ti)
			}
		}
	}
	// Owner and Creator are declared in internal/database, outside the scanned tree.
	info["Owner"] = &typeInfo{direct: true}
	info["Creator"] = &typeInfo{direct: true}

	bearing := map[string]bool{}
	for changed := true; changed; {
		changed = false
		for name, ti := range info {
			if bearing[name] {
				continue
			}
			hit := ti.direct
			for _, u := range ti.uses {
				hit = hit || bearing[u]
			}
			if hit {
				bearing[name] = true
				changed = true
			}
		}
	}
	return bearing
}

func collectStruct(st *ast.StructType, ti *typeInfo) {
	for _, f := range st.Fields.List {
		for _, n := range f.Names {
			if n.Name == "UserID" || n.Name == "CreatedBy" {
				ti.direct = true
			}
		}
		collectTypeNames(f.Type, ti)
	}
}

func collectTypeNames(e ast.Expr, ti *typeInfo) {
	switch e := e.(type) {
	case *ast.Ident:
		ti.uses = append(ti.uses, e.Name)
	case *ast.SelectorExpr:
		ti.uses = append(ti.uses, e.Sel.Name)
	case *ast.StarExpr:
		collectTypeNames(e.X, ti)
	case *ast.ArrayType:
		collectTypeNames(e.Elt, ti)
	case *ast.MapType:
		collectTypeNames(e.Key, ti)
		collectTypeNames(e.Value, ti)
	case *ast.StructType:
		collectStruct(e, ti)
	}
}

// exprBears reports whether a type expression (named or an inline struct) carries an owner.
func exprBears(e ast.Expr, owned map[string]bool) bool {
	ti := &typeInfo{}
	collectTypeNames(e, ti)
	if ti.direct {
		return true
	}
	for _, u := range ti.uses {
		if owned[u] {
			return true
		}
	}
	return false
}

// checkBinds: request bodies never bind into a type that carries an owner (BACKLOG #1). Bind into
// a request DTO and copy the allowed fields. Every bind target must be a variable whose type the
// rule can see (var v T, v := T{}, &T{}, new(T), or a parameter).
func checkBinds(fn *ast.FuncDecl, owned map[string]bool, pos func(token.Pos) string, r *ruleResult) {
	if fn.Body == nil {
		return
	}
	types := map[string]ast.Expr{}
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			for _, n := range p.Names {
				types[n.Name] = p.Type
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if n.Type != nil {
					types[name.Name] = n.Type
				} else if i < len(n.Values) {
					if t := literalType(n.Values[i]); t != nil {
						types[name.Name] = t
					}
				}
			}
		case *ast.AssignStmt:
			if n.Tok != token.DEFINE || len(n.Lhs) != len(n.Rhs) {
				return true
			}
			for i, l := range n.Lhs {
				if id, ok := l.(*ast.Ident); ok {
					if t := literalType(n.Rhs[i]); t != nil {
						types[id.Name] = t
					}
				}
			}
		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok || !bindMethods.MatchString(sel.Sel.Name) || len(n.Args) == 0 {
				return true
			}
			target := n.Args[0]
			if u, ok := target.(*ast.UnaryExpr); ok && u.Op == token.AND {
				target = u.X
			}
			id, ok := target.(*ast.Ident)
			var typ ast.Expr
			if ok {
				typ = types[id.Name]
			}
			switch {
			case typ == nil:
				r.errors = append(r.errors, fmt.Sprintf("%s: %s target: can't see its type; bind into a local `var body SomeRequest`", pos(n.Pos()), sel.Sel.Name))
			case exprBears(typ, owned):
				r.errors = append(r.errors, fmt.Sprintf("%s: %s into a type that carries user_id/created_by; bind a request DTO and copy the allowed fields (BACKLOG #1)", pos(n.Pos()), sel.Sel.Name))
			}
		}
		return true
	})
}

// literalType is the type of T{}, &T{} or new(T).
func literalType(e ast.Expr) ast.Expr {
	switch e := e.(type) {
	case *ast.CompositeLit:
		return e.Type
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return literalType(e.X)
		}
	case *ast.CallExpr:
		if id, ok := e.Fun.(*ast.Ident); ok && id.Name == "new" && len(e.Args) == 1 {
			return e.Args[0]
		}
	}
	return nil
}

// checkBothEmbedded: a struct that embeds both Owner and Creator has two BeforeCreate methods at
// the same depth, so Go promotes neither and gorm stamps nothing. A row is owned or catalog, not
// both.
func checkBothEmbedded(ts *ast.TypeSpec, pos func(token.Pos) string, r *ruleResult) {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return
	}
	embeds := map[string]bool{}
	for _, f := range st.Fields.List {
		if len(f.Names) > 0 {
			continue
		}
		ti := &typeInfo{}
		collectTypeNames(f.Type, ti)
		for _, u := range ti.uses {
			embeds[u] = true
		}
	}
	if embeds["Owner"] && embeds["Creator"] {
		r.errors = append(r.errors, fmt.Sprintf("%s: %s embeds both database.Owner and database.Creator: neither BeforeCreate hook is promoted, so nothing is stamped", pos(ts.Pos()), ts.Name.Name))
	}
}

// checkBeforeCreate: a model that embeds Owner/Creator and declares its own BeforeCreate hides
// the stamping hook unless it calls it.
func checkBeforeCreate(fn *ast.FuncDecl, owned map[string]bool, pos func(token.Pos) string, r *ruleResult) {
	if fn.Recv == nil || fn.Name.Name != "BeforeCreate" || len(fn.Recv.List) == 0 {
		return
	}
	ti := &typeInfo{}
	collectTypeNames(fn.Recv.List[0].Type, ti)
	if len(ti.uses) == 0 || !owned[ti.uses[0]] {
		return
	}
	calls := false
	ast.Inspect(fn, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "BeforeCreate" {
			if inner, ok := sel.X.(*ast.SelectorExpr); ok && (inner.Sel.Name == "Owner" || inner.Sel.Name == "Creator") {
				calls = true
			}
		}
		return true
	})
	if !calls {
		r.errors = append(r.errors, fmt.Sprintf("%s: %s.BeforeCreate hides database.Owner/Creator's stamping hook; call x.Owner.BeforeCreate(tx) (or Creator) from it", pos(fn.Pos()), ti.uses[0]))
	}
}
