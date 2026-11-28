package profile

import (
	"context"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests (Postgres only, TEST_DATABASE_URL): one profile per user, created on first save.

func profileAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterProfileRoutes(g.Group("/profile")) })
	return h, api
}

type profileJSON struct {
	HeightIn      float64 `json:"height_in"`
	Age           int     `json:"age"`
	UserID        *uint   `json:"user_id"`
	ActivityLevel string  `json:"activity_level"`
}

func getProfile(t *testing.T, api *twousers.API, u twousers.User) *profileJSON {
	t.Helper()
	var out struct{ Profile *profileJSON }
	api.Do(u, http.MethodGet, "/profile", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Profile
}

func putProfile(t *testing.T, api *twousers.API, u twousers.User, height float64, age int) {
	t.Helper()
	api.Do(u, http.MethodPut, "/profile", map[string]any{
		"height_in": height, "age": age, "sex": "male", "activity_level": "sedentary",
	}).Status(t, http.StatusOK)
}

func TestProfile_oneRowPerUserCreatedOnFirstSave(t *testing.T) {
	h, api := profileAPI(t)
	if p := getProfile(t, api, h.Alice); p != nil {
		t.Fatalf("a new user has a profile: %+v", p)
	}
	putProfile(t, api, h.Alice, 70, 30)
	if p := getProfile(t, api, h.Bob); p != nil {
		t.Fatalf("Bob sees Alice's profile: %+v", p)
	}
	putProfile(t, api, h.Bob, 64, 25)
	// Saving again updates the user's own row; it doesn't add one or touch the other's.
	putProfile(t, api, h.Bob, 65, 26)

	a, b := getProfile(t, api, h.Alice), getProfile(t, api, h.Bob)
	if a == nil || a.HeightIn != 70 || a.Age != 30 || a.UserID != nil {
		t.Fatalf("Alice's profile: %+v", a)
	}
	if b == nil || b.HeightIn != 65 || b.Age != 26 {
		t.Fatalf("Bob's profile: %+v", b)
	}
	var rows int64
	if err := h.Owner.Raw("SELECT count(*) FROM user_profiles").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 2 {
		t.Fatalf("%d profile rows, want 2", rows)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/profile", nil).Status(t, http.StatusUnauthorized)
}

func TestProfile_createIgnoresUserIDInBody(t *testing.T) {
	h, api := profileAPI(t)
	api.Do(h.Bob, http.MethodPut, "/profile", map[string]any{
		"height_in": 64, "age": 25, "sex": "female", "activity_level": "sedentary", "user_id": h.Alice.ID,
	}).Status(t, http.StatusOK)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM user_profiles").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("profile owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
	if p := getProfile(t, api, h.Alice); p != nil {
		t.Fatalf("Alice sees %+v", p)
	}
}

func TestProfile_validationErrorsAre400(t *testing.T) {
	h, api := profileAPI(t)
	api.Do(h.Bob, http.MethodPut, "/profile", map[string]any{"height_in": 0, "age": 25, "sex": "male", "activity_level": "sedentary"}).
		Status(t, http.StatusBadRequest)
	api.Do(h.Bob, http.MethodPut, "/profile", map[string]any{"height_in": 64, "age": 25, "sex": "x", "activity_level": "sedentary"}).
		Status(t, http.StatusBadRequest)
}

// A user has at most one live profile, whatever code does.
func TestProfile_databaseAllowsOneLiveRowPerUser(t *testing.T) {
	h, api := profileAPI(t)
	putProfile(t, api, h.Alice, 70, 30)
	err := h.Owner.Exec(`INSERT INTO user_profiles (height_in, age, sex, activity_level, user_id)
		VALUES (1, 1, 'male', 'sedentary', ?)`, h.Alice.ID).Error
	if err == nil {
		t.Fatal("a second live profile for Alice was accepted")
	}
}

func TestProfile_rlsBackstop(t *testing.T) {
	h, api := profileAPI(t)
	putProfile(t, api, h.Alice, 70, 30)
	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []UserProfile
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		if n := database.From(ctx).Exec("UPDATE user_profiles SET age = 99").RowsAffected; n != 0 {
			t.Errorf("unscoped update as Bob touched %d rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec(`INSERT INTO user_profiles (height_in, age, sex, activity_level, user_id)
			VALUES (1, 1, 'male', 'sedentary', ?)`, h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}
}
