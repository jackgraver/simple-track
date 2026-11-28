package grocery

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests (Postgres only, TEST_DATABASE_URL): Alice's list, Bob's token.

func groceryAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterGroceryRoutes(g.Group("/grocery")) })
	return h, api
}

type itemJSON struct {
	ID          uint    `json:"ID"`
	Name        string  `json:"name"`
	ClientID    string  `json:"client_id"`
	CompletedAt *string `json:"completed_at"`
	UserID      *uint   `json:"user_id"`
}

func addItem(t *testing.T, api *twousers.API, u twousers.User, name, clientID string) itemJSON {
	t.Helper()
	body := map[string]any{"name": name}
	if clientID != "" {
		body["client_id"] = clientID
	}
	var out struct{ Item itemJSON }
	api.Do(u, http.MethodPost, "/grocery/items", body).Status(t, http.StatusOK).JSON(t, &out)
	return out.Item
}

func listItems(t *testing.T, api *twousers.API, u twousers.User) []itemJSON {
	t.Helper()
	var out struct{ Items []itemJSON }
	api.Do(u, http.MethodGet, "/grocery/items", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Items
}

func suggestions(t *testing.T, api *twousers.API, u twousers.User) []string {
	t.Helper()
	var out struct{ Suggestions []string }
	api.Do(u, http.MethodGet, "/grocery/suggestions", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Suggestions
}

func TestGrocery_eachUserSeesOnlyTheirOwn(t *testing.T) {
	h, api := groceryAPI(t)
	addItem(t, api, h.Alice, "Milk", "")
	if got := listItems(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees %+v, want none", got)
	}
	addItem(t, api, h.Bob, "Eggs", "")
	for _, tc := range []struct {
		u    twousers.User
		want string
	}{{h.Alice, "Milk"}, {h.Bob, "Eggs"}} {
		got := listItems(t, api, tc.u)
		if len(got) != 1 || got[0].Name != tc.want || got[0].UserID != nil {
			t.Fatalf("%s sees %+v, want only %q and no user_id", tc.u.Username, got, tc.want)
		}
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/grocery/items", nil).Status(t, http.StatusUnauthorized)
}

func TestGrocery_someoneElsesIsNotFound(t *testing.T) {
	h, api := groceryAPI(t)
	const cid = "11111111-1111-4111-8111-0000000000a1"
	alices := addItem(t, api, h.Alice, "Milk", cid)
	missing := alices.ID + 1000

	// NotFound checks the exact body, so a missing item and Alice's answer the same.
	for _, ref := range []string{
		fmt.Sprint(alices.ID), fmt.Sprint(missing), cid, "11111111-1111-4111-8111-0000000000ff",
	} {
		api.Do(h.Bob, http.MethodPatch, "/grocery/items/"+ref+"/complete", nil).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, "/grocery/items/"+ref, nil).NotFound(t)
	}
	if got := listItems(t, api, h.Alice); len(got) != 1 || got[0].CompletedAt != nil {
		t.Fatalf("Alice's item was changed by Bob: %+v", got)
	}
	api.Do(h.Alice, http.MethodPatch, fmt.Sprintf("/grocery/items/%d/complete", alices.ID), nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, "/grocery/items/"+cid, nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, "/grocery/items/"+cid, nil).NotFound(t)
}

// The same client_id is two different items for two users, and each user's replay is idempotent.
func TestGrocery_sameClientIDForTwoUsers(t *testing.T) {
	h, api := groceryAPI(t)
	const cid = "11111111-1111-4111-8111-0000000000b2"
	a := addItem(t, api, h.Alice, "Milk", cid)
	b := addItem(t, api, h.Bob, "Eggs", cid)
	if a.ID == b.ID || b.Name != "Eggs" {
		t.Fatalf("Bob's create with Alice's client_id returned %+v (Alice's is %+v)", b, a)
	}
	if again := addItem(t, api, h.Bob, "Eggs", cid); again.ID != b.ID {
		t.Fatalf("Bob's replay created a new row: %d vs %d", again.ID, b.ID)
	}
	api.Do(h.Bob, http.MethodPatch, "/grocery/items/"+cid+"/complete", nil).Status(t, http.StatusOK)
	if got := listItems(t, api, h.Alice); len(got) != 1 || got[0].CompletedAt != nil {
		t.Fatalf("Bob completing by client_id touched Alice's item: %+v", got)
	}
}

// Suggestions (history of completed items), parse matching and bulk add are per user.
func TestGrocery_suggestionsParseAndBulkArePerUser(t *testing.T) {
	h, api := groceryAPI(t)
	milk := addItem(t, api, h.Alice, "Milk", "")
	api.Do(h.Alice, http.MethodPatch, fmt.Sprintf("/grocery/items/%d/complete", milk.ID), nil).Status(t, http.StatusOK)
	addItem(t, api, h.Alice, "Bread", "")

	if got := suggestions(t, api, h.Alice); len(got) != 1 || got[0] != "Milk" {
		t.Fatalf("Alice's suggestions: %v", got)
	}
	if got := suggestions(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees Alice's history: %v", got)
	}

	var parsed struct{ Lines []ParsedLine }
	api.Do(h.Bob, http.MethodPost, "/grocery/items/parse", map[string]any{"text": "milk\nbread"}).Status(t, http.StatusOK).JSON(t, &parsed)
	for _, l := range parsed.Lines {
		if l.Matched || l.OnList {
			t.Fatalf("Bob's parse matched Alice's items: %+v", parsed.Lines)
		}
	}

	// Bob adds "Bread" although it is on Alice's list; it isn't skipped.
	var bulk struct {
		Items   []itemJSON
		Skipped []string
	}
	api.Do(h.Bob, http.MethodPost, "/grocery/items/bulk", map[string]any{"names": []string{"Bread", "Bread"}}).
		Status(t, http.StatusOK).JSON(t, &bulk)
	if len(bulk.Items) != 1 || len(bulk.Skipped) != 1 {
		t.Fatalf("bulk: %+v", bulk)
	}
	if got := listItems(t, api, h.Alice); len(got) != 1 || got[0].Name != "Bread" {
		t.Fatalf("Alice's list changed: %+v", got)
	}
}

func TestGrocery_createIgnoresUserIDInBody(t *testing.T) {
	h, api := groceryAPI(t)
	api.Do(h.Bob, http.MethodPost, "/grocery/items", map[string]any{"name": "Eggs", "user_id": h.Alice.ID}).Status(t, http.StatusOK)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM grocery_items").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("item owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
}

func TestGrocery_rlsBackstop(t *testing.T) {
	h, api := groceryAPI(t)
	addItem(t, api, h.Alice, "Milk", "")
	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []GroceryItem
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		if n := database.From(ctx).Exec("DELETE FROM grocery_items").RowsAffected; n != 0 {
			t.Errorf("unscoped delete as Bob touched %d rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO grocery_items (name, user_id) VALUES ('x', ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}
}
