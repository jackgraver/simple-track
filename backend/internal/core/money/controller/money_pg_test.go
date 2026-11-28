package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"be-simpletracker/internal/core/money/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests for the money module (docs/multi-user/DECISIONS.md, "C4 money"): accounts,
// balances and deposits are owned; contribution types and their limits are a shared catalog; the
// room figures are per user. Postgres only (TEST_DATABASE_URL).

func moneyAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterRoutes(g.Group("/money")) })
	return h, api
}

type accountJSON struct {
	ID             uint     `json:"ID"`
	Name           string   `json:"name"`
	CurrentBalance float64  `json:"current_balance"`
	UserID         *uint    `json:"user_id"`
	TypeID         *uint    `json:"contribution_type_id"`
	TotalDeposits  float64  `json:"total_deposits"`
	CreatedBy      *uint    `json:"created_by"`
	Unused         struct{} `json:"-"`
}

func createAccount(t *testing.T, api *twousers.API, u twousers.User, body map[string]any) accountJSON {
	t.Helper()
	var out struct{ Account accountJSON }
	api.Do(u, http.MethodPost, "/money/accounts", body).Status(t, http.StatusCreated).JSON(t, &out)
	return out.Account
}

func listAccounts(t *testing.T, api *twousers.API, u twousers.User) []accountJSON {
	t.Helper()
	var out struct{ Accounts []accountJSON }
	api.Do(u, http.MethodGet, "/money/accounts?archived=true", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Accounts
}

type netWorthPoint struct {
	Assets      float64 `json:"assets"`
	Liabilities float64 `json:"liabilities"`
	NetWorth    float64 `json:"net_worth"`
}

func netWorth(t *testing.T, api *twousers.API, u twousers.User) []netWorthPoint {
	t.Helper()
	var out struct{ Points []netWorthPoint }
	api.Do(u, http.MethodGet, "/money/net-worth/history", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Points
}

func accountPath(id uint, rest string) string { return fmt.Sprintf("/money/accounts/%d%s", id, rest) }

// alicesAccount creates a brokerage account for Alice with a balance and a deposit.
func alicesAccount(t *testing.T, h *twousers.Harness, api *twousers.API) (accountJSON, uint, uint) {
	t.Helper()
	acct := createAccount(t, api, h.Alice, map[string]any{
		"name": "Alice TFSA", "kind": "investment", "opening_balance": 1000, "balance_date": "2026-01-10",
	})
	var tx struct{ Transaction struct{ ID uint } }
	api.Do(h.Alice, http.MethodPost, accountPath(acct.ID, "/transactions"),
		map[string]any{"kind": "deposit", "amount": 250, "date": "2026-01-11"}).Status(t, http.StatusCreated).JSON(t, &tx)
	var bal struct{ Balances []struct{ ID uint } }
	api.Do(h.Alice, http.MethodGet, accountPath(acct.ID, "/balances"), nil).Status(t, http.StatusOK).JSON(t, &bal)
	if len(bal.Balances) != 1 {
		t.Fatalf("Alice's balances = %+v", bal.Balances)
	}
	return acct, bal.Balances[0].ID, tx.Transaction.ID
}

func TestMoney_eachUserSeesOnlyTheirOwnAccounts(t *testing.T) {
	h, api := moneyAPI(t)
	alicesAccount(t, h, api)
	// Account names are per user: Bob can have his own "Alice TFSA"-style name too.
	createAccount(t, api, h.Bob, map[string]any{"name": "Alice TFSA", "kind": "savings", "opening_balance": 5, "balance_date": "2026-01-10"})

	for _, u := range []twousers.User{h.Alice, h.Bob} {
		got := listAccounts(t, api, u)
		if len(got) != 1 || got[0].Name != "Alice TFSA" {
			t.Fatalf("%s sees %+v, want exactly their own account", u.Username, got)
		}
		if got[0].UserID != nil {
			t.Fatalf("user_id is in the response: %+v", got[0])
		}
	}
	if bob := listAccounts(t, api, h.Bob); bob[0].CurrentBalance != 5 {
		t.Fatalf("Bob sees balance %v, want his own 5", bob[0].CurrentBalance)
	}
	if got := listAccounts(t, api, h.Admin); len(got) != 0 {
		t.Fatalf("admin sees %+v, want none: accounts are personal even for the admin", got)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/money/accounts", nil).Status(t, http.StatusUnauthorized)
}

func TestMoney_someoneElsesAccountIsNotFound(t *testing.T) {
	h, api := moneyAPI(t)
	acct, balanceID, txID := alicesAccount(t, h, api)
	missing := acct.ID + 1000
	date := map[string]any{"date": "2026-02-01", "balance": 1}

	for _, id := range []uint{acct.ID, missing} {
		api.Do(h.Bob, http.MethodGet, accountPath(id, ""), nil).NotFound(t)
		api.Do(h.Bob, http.MethodPut, accountPath(id, ""), map[string]any{"name": "Mine now", "kind": "cash"}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, accountPath(id, ""), nil).NotFound(t)

		api.Do(h.Bob, http.MethodGet, accountPath(id, "/balances"), nil).NotFound(t)
		api.Do(h.Bob, http.MethodPut, accountPath(id, "/balances"), date).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, accountPath(id, fmt.Sprintf("/balances/%d", balanceID)), nil).NotFound(t)

		api.Do(h.Bob, http.MethodGet, accountPath(id, "/transactions"), nil).NotFound(t)
		api.Do(h.Bob, http.MethodPost, accountPath(id, "/transactions"),
			map[string]any{"kind": "deposit", "amount": 5, "date": "2026-02-01", "update_balance": true}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, accountPath(id, fmt.Sprintf("/transactions/%d", txID)), nil).NotFound(t)
	}

	// The bulk balance update answers the same for Alice's account as for a missing one.
	bulk := func(id uint) *twousers.Response {
		return api.Do(h.Bob, http.MethodPut, "/money/balances", map[string]any{
			"date": "2026-02-01", "balances": []map[string]any{{"account_id": id, "balance": 9}},
		}).Status(t, http.StatusBadRequest)
	}
	// (The message names the requested ID, so mask it before comparing.)
	mask := func(r *twousers.Response, id uint) string {
		return strings.ReplaceAll(string(r.Body), fmt.Sprint(id), "N")
	}
	if a, m := mask(bulk(acct.ID), acct.ID), mask(bulk(missing), missing); a != m {
		t.Fatalf("Alice's account answers %s, a missing one %s: must be the same", a, m)
	}

	// Nothing changed for Alice, and her own calls still work.
	got := listAccounts(t, api, h.Alice)
	if len(got) != 1 || got[0].Name != "Alice TFSA" || got[0].CurrentBalance != 1000 || got[0].TotalDeposits != 250 {
		t.Fatalf("Alice's account changed: %+v", got)
	}
	api.Do(h.Alice, http.MethodPut, accountPath(acct.ID, "/balances"), date).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, accountPath(acct.ID, fmt.Sprintf("/transactions/%d", txID)), nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, accountPath(acct.ID, ""), nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, accountPath(acct.ID, ""), nil).NotFound(t)
}

// A forged user_id in any body is ignored: the row belongs to whoever is signed in.
func TestMoney_createIgnoresUserIDInBody(t *testing.T) {
	h, api := moneyAPI(t)
	acct := createAccount(t, api, h.Bob, map[string]any{
		"name": "Sneaky", "kind": "cash", "user_id": h.Alice.ID, "opening_balance": 1, "balance_date": "2026-01-10",
	})
	api.Do(h.Bob, http.MethodPut, accountPath(acct.ID, "/balances"),
		map[string]any{"date": "2026-02-01", "balance": 2, "user_id": h.Alice.ID}).Status(t, http.StatusOK)
	api.Do(h.Bob, http.MethodPost, accountPath(acct.ID, "/transactions"),
		map[string]any{"kind": "deposit", "amount": 3, "date": "2026-02-01", "user_id": h.Alice.ID}).Status(t, http.StatusCreated)

	for _, table := range []string{"investment_accounts", "account_balances", "investment_deposits"} {
		var wrong int64
		if err := h.Owner.Raw("SELECT count(*) FROM "+table+" WHERE user_id <> ?", h.Bob.ID).Scan(&wrong).Error; err != nil {
			t.Fatal(err)
		}
		if wrong != 0 {
			t.Fatalf("%s has %d rows not owned by Bob", table, wrong)
		}
	}
	if got := listAccounts(t, api, h.Alice); len(got) != 0 {
		t.Fatalf("Alice sees %+v", got)
	}
}

func TestMoney_netWorthIsPerUser(t *testing.T) {
	h, api := moneyAPI(t)
	alicesAccount(t, h, api) // 1000 on Jan 10
	createAccount(t, api, h.Bob, map[string]any{"name": "Bob chequing", "kind": "chequing", "opening_balance": 40, "balance_date": "2026-01-10"})
	createAccount(t, api, h.Bob, map[string]any{"name": "Bob card", "kind": "credit_card", "opening_balance": 15, "balance_date": "2026-01-10"})

	alice, bob := netWorth(t, api, h.Alice), netWorth(t, api, h.Bob)
	if len(alice) != 1 || alice[0].NetWorth != 1000 || alice[0].Assets != 1000 {
		t.Fatalf("Alice's net worth = %+v, want 1000", alice)
	}
	if len(bob) != 1 || bob[0].Assets != 40 || bob[0].Liabilities != 15 || bob[0].NetWorth != 25 {
		t.Fatalf("Bob's net worth = %+v, want assets 40, liabilities 15, net 25 (nothing of Alice's)", bob)
	}
	if got := netWorth(t, api, h.Admin); len(got) != 0 {
		t.Fatalf("admin's net worth = %+v, want none", got)
	}
}

// Row-level security is the backstop: queries that forget database.Owned still see only the
// caller's rows, and writes that point at or for someone else are refused by the database.
func TestMoney_rlsBackstop(t *testing.T) {
	h, api := moneyAPI(t)
	acct, _, _ := alicesAccount(t, h, api)

	err := h.Run(h.Bob, func(ctx context.Context) error {
		db := database.From(ctx)
		var accounts []models.Account
		var balances []models.AccountBalance
		var deposits []models.AccountTransaction
		var rooms []models.ContributionRoom
		for _, dest := range []any{&accounts, &balances, &deposits, &rooms} {
			if err := db.Unscoped().Find(dest).Error; err != nil { // no Owned scope
				return err
			}
		}
		if len(accounts)+len(balances)+len(deposits)+len(rooms) != 0 {
			t.Errorf("unscoped queries as Bob returned Alice's rows: %d accounts, %d balances, %d deposits, %d rooms",
				len(accounts), len(balances), len(deposits), len(rooms))
		}
		for _, table := range []string{"investment_accounts", "account_balances", "investment_deposits"} {
			if n := db.Exec("UPDATE " + table + " SET updated_at = now()").RowsAffected; n != 0 {
				t.Errorf("unscoped update of %s as Bob touched %d rows", table, n)
			}
			if n := db.Exec("DELETE FROM " + table).RowsAffected; n != 0 {
				t.Errorf("unscoped delete from %s as Bob removed %d rows", table, n)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Bob can't insert rows for Alice, nor hang his own rows off her account.
	for name, stmt := range map[string]string{
		"account for Alice":          "INSERT INTO investment_accounts (name, kind, user_id, current_balance) VALUES ('x', 'cash', %[1]d, 0)",
		"deposit on Alice's account": "INSERT INTO investment_deposits (account_id, user_id, kind, amount, date) VALUES (%[2]d, %[3]d, 'deposit', 1, now())",
		"balance on Alice's account": "INSERT INTO account_balances (account_id, user_id, date, balance) VALUES (%[2]d, %[3]d, now(), 1)",
	} {
		err := h.Run(h.Bob, func(ctx context.Context) error {
			return database.From(ctx).Exec(fmt.Sprintf(stmt, h.Alice.ID, acct.ID, h.Bob.ID)).Error
		})
		if err == nil {
			t.Errorf("%s as Bob succeeded; want a row-level security or foreign key error", name)
		}
	}

	var n int64
	if err := h.App.Model(&models.Account{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("app role without app.user_id sees %d accounts", n)
	}
	_ = api
}

// ---- Shared catalog ----

type typeJSON struct {
	ID              uint     `json:"ID"`
	Name            string   `json:"name"`
	CreatedBy       uint     `json:"created_by"`
	CanEdit         bool     `json:"can_edit"`
	EligibilityAge  *int     `json:"eligibility_age"`
	CarryForwardCap *float64 `json:"carry_forward_cap"`
	StartYear       *int     `json:"start_year"`
	OpeningRoomYear *int     `json:"opening_room_year"`
	OpeningRoom     *float64 `json:"opening_room"`
	Limits          []struct {
		Year        int     `json:"year"`
		AnnualLimit float64 `json:"annual_limit"`
	} `json:"limits"`
	Room *struct {
		Available           float64 `json:"available"`
		ContributedThisYear float64 `json:"contributed_this_year"`
	} `json:"room"`
	RoomIssue string `json:"room_issue"`
}

func createType(t *testing.T, api *twousers.API, u twousers.User, body map[string]any) typeJSON {
	t.Helper()
	var out struct {
		Type typeJSON `json:"contribution_type"`
	}
	api.Do(u, http.MethodPost, "/money/contribution-types", body).Status(t, http.StatusCreated).JSON(t, &out)
	return out.Type
}

func listTypes(t *testing.T, api *twousers.API, u twousers.User) []typeJSON {
	t.Helper()
	var out struct {
		Types []typeJSON `json:"contribution_types"`
	}
	api.Do(u, http.MethodGet, "/money/contribution-types", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Types
}

func typePath(id uint, rest string) string {
	return fmt.Sprintf("/money/contribution-types/%d%s", id, rest)
}

func findType(t *testing.T, types []typeJSON, id uint) typeJSON {
	t.Helper()
	for _, ty := range types {
		if ty.ID == id {
			return ty
		}
	}
	t.Fatalf("type %d not in %+v", id, types)
	return typeJSON{}
}

func TestMoneyCatalog_anyoneAddsCreatorOrAdminEditsOnlyAdminDeletes(t *testing.T) {
	h, api := moneyAPI(t)
	bobs := createType(t, api, h.Bob, map[string]any{"name": "Bob's plan", "carry_forward_cap": 100})
	if bobs.CreatedBy != h.Bob.ID {
		t.Fatalf("created_by = %d, want Bob (%d)", bobs.CreatedBy, h.Bob.ID)
	}
	api.Do(h.Bob, http.MethodPut, typePath(bobs.ID, "/limits/2026"), map[string]any{"annual_limit": 500}).Status(t, http.StatusOK)

	// Everyone signed in sees it; only its creator (and the admin) is offered edit.
	for u, canEdit := range map[twousers.User]bool{h.Alice: false, h.Bob: true, h.Admin: true} {
		got := findType(t, listTypes(t, api, u), bobs.ID)
		if got.CanEdit != canEdit || len(got.Limits) != 1 {
			t.Fatalf("%s sees %+v, want can_edit=%v and 1 limit", u.Username, got, canEdit)
		}
	}

	// Alice can't edit it, add or change its limits, or delete it: 403, and nothing changes.
	api.Do(h.Alice, http.MethodPut, typePath(bobs.ID, ""), map[string]any{"name": "Alice's now"}).Status(t, http.StatusForbidden)
	api.Do(h.Alice, http.MethodPut, typePath(bobs.ID, "/limits/2027"), map[string]any{"annual_limit": 1}).Status(t, http.StatusForbidden)
	api.Do(h.Alice, http.MethodPut, typePath(bobs.ID, "/limits/2026"), map[string]any{"annual_limit": 1}).Status(t, http.StatusForbidden)
	api.Do(h.Alice, http.MethodDelete, typePath(bobs.ID, "/limits/2026"), nil).Status(t, http.StatusForbidden)
	api.Do(h.Alice, http.MethodDelete, typePath(bobs.ID, ""), nil).Status(t, http.StatusForbidden)
	got := findType(t, listTypes(t, api, h.Bob), bobs.ID)
	if got.Name != "Bob's plan" || len(got.Limits) != 1 || got.Limits[0].AnnualLimit != 500 {
		t.Fatalf("Alice changed Bob's type: %+v", got)
	}
	// A type that doesn't exist is a 404 (nothing to be forbidden from)...
	api.Do(h.Alice, http.MethodPut, typePath(bobs.ID+1000, ""), map[string]any{"name": "x"}).NotFound(t)
	api.Do(h.Alice, http.MethodPut, typePath(bobs.ID+1000, "/room"), map[string]any{}).NotFound(t)

	// ...the creator edits, but doesn't delete.
	api.Do(h.Bob, http.MethodPut, typePath(bobs.ID, ""), map[string]any{"name": "Bob's plan v2"}).Status(t, http.StatusOK)
	api.Do(h.Bob, http.MethodDelete, typePath(bobs.ID, ""), nil).Status(t, http.StatusForbidden)

	// The admin edits anyone's type, changes its limits and deletes it.
	api.Do(h.Admin, http.MethodPut, typePath(bobs.ID, ""), map[string]any{"name": "Bob's plan v3"}).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodPut, typePath(bobs.ID, "/limits/2026"), map[string]any{"annual_limit": 600}).Status(t, http.StatusOK)
	// Bob (the type's creator) can still change a limit the admin touched.
	api.Do(h.Bob, http.MethodPut, typePath(bobs.ID, "/limits/2026"), map[string]any{"annual_limit": 700}).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodDelete, typePath(bobs.ID, "/limits/2026"), nil).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodDelete, typePath(bobs.ID, ""), nil).Status(t, http.StatusOK)
	if got := listTypes(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("type still listed after the admin deleted it: %+v", got)
	}
}

// The database enforces the same rules when a query forgets the handler's checks.
func TestMoneyCatalog_rlsBackstop(t *testing.T) {
	h, api := moneyAPI(t)
	bobs := createType(t, api, h.Bob, map[string]any{"name": "Bob's plan"})
	api.Do(h.Bob, http.MethodPut, typePath(bobs.ID, "/limits/2026"), map[string]any{"annual_limit": 500}).Status(t, http.StatusOK)

	exec := func(u twousers.User, sql string) (int64, error) {
		var n int64
		err := h.Run(u, func(ctx context.Context) error {
			res := database.From(ctx).Exec(sql)
			n = res.RowsAffected
			return res.Error
		})
		return n, err
	}
	for _, table := range []string{"investment_account_types", "investment_contribution_rules"} {
		// Alice's raw update of Bob's rows matches nothing; her delete is refused or matches nothing.
		if n, err := exec(h.Alice, "UPDATE "+table+" SET updated_at = now()"); err != nil || n != 0 {
			t.Errorf("%s: Alice updated %d of Bob's rows (err %v)", table, n, err)
		}
		if n, _ := exec(h.Alice, "DELETE FROM "+table); n != 0 {
			t.Errorf("%s: Alice deleted %d of Bob's rows", table, n)
		}
		// The soft delete (an UPDATE of deleted_at) is the admin's alone, even for the creator.
		if _, err := exec(h.Bob, "UPDATE "+table+" SET deleted_at = now()"); err == nil {
			t.Errorf("%s: Bob soft-deleted a catalog row", table)
		}
	}
	if _, err := exec(h.Alice, fmt.Sprintf(
		"INSERT INTO investment_account_types (name, created_by, withdrawals_restore_room) VALUES ('x', %d, false)", h.Bob.ID)); err == nil {
		t.Error("Alice added a catalog row as Bob")
	}
	if _, err := exec(h.Bob, "UPDATE investment_account_types SET created_by = "+fmt.Sprint(h.Alice.ID)); err == nil {
		t.Error("Bob changed created_by")
	}
	// Without an identity the catalog is empty too.
	var n int64
	if err := h.App.Model(&models.ContributionType{}).Count(&n).Error; err != nil || n != 0 {
		t.Fatalf("app role without app.user_id sees %d types (err %v)", n, err)
	}
}

func setRoom(t *testing.T, api *twousers.API, u twousers.User, id uint, body map[string]any) typeJSON {
	t.Helper()
	var out struct {
		Type typeJSON `json:"contribution_type"`
	}
	api.Do(u, http.MethodPut, typePath(id, "/room"), body).Status(t, http.StatusOK).JSON(t, &out)
	return out.Type
}

// Room figures and the flows behind them are per user, on a type both share.
func TestMoneyCatalog_roomIsPerUser(t *testing.T) {
	h, api := moneyAPI(t)
	year := time.Now().Year() // the room maths runs on the real clock
	var out struct {
		Type typeJSON `json:"contribution_type"`
	}
	api.Do(h.Alice, http.MethodPost, "/money/contribution-types/presets", map[string]any{"key": "tfsa", "opening_room_year": year, "opening_room": 20000}).
		Status(t, http.StatusCreated).JSON(t, &out)
	tfsa := out.Type

	// Bob picks the same preset: the shared TFSA is reused, not duplicated, and he gets his own room.
	api.Do(h.Bob, http.MethodPost, "/money/contribution-types/presets", map[string]any{"key": "tfsa", "opening_room_year": year, "opening_room": 3000}).
		Status(t, http.StatusCreated).JSON(t, &out)
	if out.Type.ID != tfsa.ID {
		t.Fatalf("Bob's TFSA is type %d, want Alice's %d", out.Type.ID, tfsa.ID)
	}
	if types := listTypes(t, api, h.Bob); len(types) != 1 {
		t.Fatalf("%d types after both added a TFSA, want 1", len(types))
	}
	// A manual create with a taken name is still refused.
	api.Do(h.Bob, http.MethodPost, "/money/contribution-types", map[string]any{"name": "tfsa"}).Status(t, http.StatusBadRequest)

	// Both have an account called "TFSA" on that type; only Alice deposits.
	for _, u := range []twousers.User{h.Alice, h.Bob} {
		createAccount(t, api, u, map[string]any{"name": "TFSA", "kind": "investment", "contribution_type_id": tfsa.ID})
	}
	alicesTFSA := listAccounts(t, api, h.Alice)[0]
	api.Do(h.Alice, http.MethodPost, accountPath(alicesTFSA.ID, "/transactions"),
		map[string]any{"kind": "deposit", "amount": 4000, "date": fmt.Sprintf("%d-01-02", year)}).Status(t, http.StatusCreated)

	aliceType := findType(t, listTypes(t, api, h.Alice), tfsa.ID)
	bobType := findType(t, listTypes(t, api, h.Bob), tfsa.ID)
	if aliceType.OpeningRoom == nil || *aliceType.OpeningRoom != 20000 || bobType.OpeningRoom == nil || *bobType.OpeningRoom != 3000 {
		t.Fatalf("opening rooms: Alice %v, Bob %v, want their own 20000 and 3000", aliceType.OpeningRoom, bobType.OpeningRoom)
	}
	if aliceType.Room == nil || bobType.Room == nil {
		t.Fatalf("room issues: Alice %q, Bob %q", aliceType.RoomIssue, bobType.RoomIssue)
	}
	if aliceType.Room.ContributedThisYear != 4000 || aliceType.Room.Available != 16000 {
		t.Fatalf("Alice's room = %+v, want 4000 contributed of 20000", aliceType.Room)
	}
	if bobType.Room.ContributedThisYear != 0 || bobType.Room.Available != 3000 {
		t.Fatalf("Bob's room = %+v, want his own 3000 and none of Alice's deposit", bobType.Room)
	}

	// Bob changes his room; Alice's is untouched. Clearing it falls back to the rules.
	got := setRoom(t, api, h.Bob, tfsa.ID, map[string]any{"opening_room_year": 2026, "opening_room": 1234.5})
	if got.OpeningRoom == nil || *got.OpeningRoom != 1234.5 {
		t.Fatalf("Bob's room after saving = %+v", got.OpeningRoom)
	}
	if a := findType(t, listTypes(t, api, h.Alice), tfsa.ID); a.OpeningRoom == nil || *a.OpeningRoom != 20000 {
		t.Fatalf("Bob's change reached Alice: %+v", a.OpeningRoom)
	}
	if got := setRoom(t, api, h.Bob, tfsa.ID, map[string]any{}); got.OpeningRoom != nil {
		t.Fatalf("cleared room still %v", *got.OpeningRoom)
	}
	// Bob can't edit the shared rules of Alice's type, but his own room is always his to set.
	api.Do(h.Bob, http.MethodPut, typePath(tfsa.ID, ""), map[string]any{"name": "TFSA"}).Status(t, http.StatusForbidden)

	var rooms int64
	if err := h.Owner.Raw("SELECT count(*) FROM investment_contribution_room WHERE user_id = ?", h.Bob.ID).Scan(&rooms).Error; err != nil || rooms != 0 {
		t.Fatalf("Bob has %d room rows after clearing (err %v)", rooms, err)
	}
	if err := h.Owner.Raw("SELECT count(*) FROM investment_contribution_room WHERE user_id = ?", h.Alice.ID).Scan(&rooms).Error; err != nil || rooms != 1 {
		t.Fatalf("Alice has %d room rows (err %v)", rooms, err)
	}
	if strings.Contains(string(api.Do(h.Bob, http.MethodGet, "/money/contribution-types", nil).Body), "user_id") {
		t.Fatal("user_id leaked in the contribution types response")
	}
}

// The admin's delete of a type other users' accounts still use is refused, not a 500.
func TestMoneyCatalog_deleteRefusedWhileOthersUseIt(t *testing.T) {
	h, api := moneyAPI(t)
	bobs := createType(t, api, h.Bob, map[string]any{"name": "Bob's plan"})
	createAccount(t, api, h.Bob, map[string]any{"name": "Plan account", "kind": "investment", "contribution_type_id": bobs.ID})
	api.Do(h.Admin, http.MethodDelete, typePath(bobs.ID, ""), nil).Status(t, http.StatusBadRequest)
	if got := listTypes(t, api, h.Bob); len(got) != 1 {
		t.Fatalf("type gone: %+v", got)
	}
}
