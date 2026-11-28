package diet_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"be-simpletracker/internal/core/diet"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests for diet (docs/multi-user/DECISIONS.md "C2 diet"): foods are a shared catalog;
// days, meals, saved meals, recipes, plans and everything under them are each user's own.
// Postgres only (TEST_DATABASE_URL), so row-level security applies.

func dietAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	return h, h.Router(func(g *gin.RouterGroup) { diet.RegisterGroup(g.Group("/diet")) })
}

type idJSON struct {
	ID uint `json:"ID"`
}

func do(t *testing.T, api *twousers.API, u twousers.User, method, path string, body any, want int) *twousers.Response {
	t.Helper()
	return api.Do(u, method, path, body).Status(t, want)
}

// food creates a catalog food as u and returns its ID.
func food(t *testing.T, api *twousers.API, u twousers.User, name string, extra map[string]any) uint {
	t.Helper()
	body := map[string]any{"name": name, "serving_type": "g", "serving_amount": 100, "calories": 100, "protein": 10}
	for k, v := range extra {
		body[k] = v
	}
	var out struct {
		Food struct {
			ID        uint `json:"ID"`
			CreatedBy uint `json:"created_by"`
		} `json:"food"`
	}
	do(t, api, u, http.MethodPost, "/diet/foods", body, http.StatusCreated).JSON(t, &out)
	return out.Food.ID
}

func mealBody(name string, foodID uint, extra map[string]any) map[string]any {
	m := map[string]any{"name": name, "items": []map[string]any{{"food_id": foodID, "amount": 2}}}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// newMeal creates a meal (not logged) as u and returns its ID.
func newMeal(t *testing.T, api *twousers.API, u twousers.User, name string, foodID uint) uint {
	t.Helper()
	var out struct {
		MealID uint `json:"meal_id"`
	}
	do(t, api, u, http.MethodPost, "/diet/meals/meal/new", map[string]any{"meal": mealBody(name, foodID, nil)}, http.StatusCreated).JSON(t, &out)
	return out.MealID
}

// today is the caller's diet day for today (created on first read) and its plan.
func today(t *testing.T, api *twousers.API, u twousers.User) (dayID, planID uint) {
	t.Helper()
	var out struct {
		Day struct {
			ID     uint `json:"ID"`
			PlanID uint `json:"plan_id"`
		} `json:"day"`
	}
	do(t, api, u, http.MethodGet, "/diet/logs/today", nil, http.StatusOK).JSON(t, &out)
	return out.Day.ID, out.Day.PlanID
}

func foodNames(t *testing.T, api *twousers.API, u twousers.User) (foods []string, composites []string) {
	t.Helper()
	var out struct {
		Foods []struct {
			Name string `json:"name"`
		} `json:"foods"`
		Composite []struct {
			Name string `json:"name"`
		} `json:"composite_foods"`
	}
	do(t, api, u, http.MethodGet, "/diet/meals/food/all", nil, http.StatusOK).JSON(t, &out)
	for _, f := range out.Foods {
		foods = append(foods, f.Name)
	}
	for _, c := range out.Composite {
		composites = append(composites, c.Name)
	}
	return foods, composites
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func TestDiet_foodCatalog(t *testing.T) {
	h, api := dietAPI(t)

	// Anyone adds; created_by is the caller whatever the body says.
	var created struct {
		Food struct {
			ID        uint `json:"ID"`
			CreatedBy uint `json:"created_by"`
		} `json:"food"`
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/foods", map[string]any{
		"name": "Oats", "serving_type": "g", "serving_amount": 40, "calories": 150,
		"created_by": h.Bob.ID, "user_id": h.Bob.ID,
	}, http.StatusCreated).JSON(t, &created)
	if created.Food.CreatedBy != h.Alice.ID {
		t.Fatalf("created_by %d, want Alice (%d): a forged value must be ignored", created.Food.CreatedBy, h.Alice.ID)
	}
	oats := created.Food.ID

	// Everyone reads it.
	if foods, _ := foodNames(t, api, h.Bob); !contains(foods, "Oats") {
		t.Fatalf("Bob's picker %v lacks Alice's food", foods)
	}

	// Only the creator or an admin edits; the row is visible, so it's a 403, and a missing one a 404.
	edit := map[string]any{"name": "Rolled oats", "serving_type": "g", "serving_amount": 40, "calories": 160}
	do(t, api, h.Bob, http.MethodPut, fmt.Sprintf("/diet/foods/%d", oats), edit, http.StatusForbidden)
	do(t, api, h.Bob, http.MethodPut, "/diet/foods/999999", edit, http.StatusNotFound)
	do(t, api, h.Alice, http.MethodPut, fmt.Sprintf("/diet/foods/%d", oats), edit, http.StatusOK)
	edit["name"] = "Oats (admin)"
	do(t, api, h.Admin, http.MethodPut, fmt.Sprintf("/diet/foods/%d", oats), edit, http.StatusOK)
	var calories float64
	if err := h.Owner.Raw("SELECT calories FROM foods WHERE id = ?", oats).Scan(&calories).Error; err != nil || calories != 160 {
		t.Fatalf("calories %v (%v), want the edit's 160", calories, err)
	}

	// Renaming onto another food's name is a 409, not a 500.
	food(t, api, h.Bob, "Rice", nil)
	edit["name"] = "Rice"
	do(t, api, h.Alice, http.MethodPut, fmt.Sprintf("/diet/foods/%d", oats), edit, http.StatusConflict)

	// Only the admin deletes. Creator and others get 403, and the food is still listed.
	do(t, api, h.Alice, http.MethodDelete, fmt.Sprintf("/diet/foods/%d", oats), nil, http.StatusForbidden)
	do(t, api, h.Bob, http.MethodDelete, fmt.Sprintf("/diet/foods/%d", oats), nil, http.StatusForbidden)
	do(t, api, h.Bob, http.MethodDelete, "/diet/foods/999999", nil, http.StatusNotFound)
	if foods, _ := foodNames(t, api, h.Alice); !contains(foods, "Oats (admin)") {
		t.Fatalf("food vanished after refused deletes: %v", foods)
	}
	// A saved meal using it before the delete.
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{
		"name": "Breakfast", "items": []map[string]any{{"food_id": oats, "amount": 1}}}, http.StatusCreated)
	do(t, api, h.Admin, http.MethodDelete, fmt.Sprintf("/diet/foods/%d", oats), nil, http.StatusNoContent)
	var saved struct {
		Saved []struct {
			ID    uint `json:"ID"`
			Items []struct {
				Food struct{ Name string } `json:"food"`
			} `json:"items"`
		} `json:"saved_meals"`
	}
	do(t, api, h.Alice, http.MethodGet, "/diet/meals/saved-meal/all", nil, http.StatusOK).JSON(t, &saved)
	if len(saved.Saved) != 1 || saved.Saved[0].Items[0].Food.Name == "" {
		t.Fatalf("history lost the deleted food: %+v", saved.Saved)
	}
	// ...and it can still be planned and edited.
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": saved.Saved[0].ID}, http.StatusOK)
	do(t, api, h.Alice, http.MethodPut, fmt.Sprintf("/diet/meals/saved-meal/%d", saved.Saved[0].ID), map[string]any{
		"name": "Breakfast", "items": []map[string]any{{"food_id": oats, "amount": 2}}}, http.StatusOK)
	if foods, _ := foodNames(t, api, h.Bob); contains(foods, "Oats (admin)") {
		t.Fatalf("deleted food still in the picker: %v", foods)
	}
}

// A body that carries a nested food never creates or edits a catalog food.
func TestDiet_nestedFoodIsIgnored(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Alice, "Rice", nil)
	nested := map[string]any{"ID": 0, "name": "Smuggled", "calories": 1, "serving_type": "g", "serving_amount": 1}
	item := map[string]any{"food_id": rice, "amount": 1, "food": nested}

	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/new", map[string]any{"meal": map[string]any{"name": "Bowl", "items": []any{item}}}, http.StatusCreated)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{"name": "Bowl", "items": []any{item}}, http.StatusCreated)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/composite-food/new", map[string]any{"name": "Mix", "items": []any{item}}, http.StatusCreated)

	var n int64
	if err := h.Owner.Raw("SELECT count(*) FROM foods WHERE name = 'Smuggled'").Scan(&n).Error; err != nil || n != 0 {
		t.Fatalf("%d foods created through a nested body (%v)", n, err)
	}
	var cal float64
	if err := h.Owner.Raw("SELECT calories FROM foods WHERE id = ?", rice).Scan(&cal).Error; err != nil || cal != 100 {
		t.Fatalf("nested food edited Alice's: calories %v (%v)", cal, err)
	}
}

func TestDiet_variantLinks(t *testing.T) {
	h, api := dietAPI(t)
	lean := food(t, api, h.Alice, "Ground beef 90/10", nil)

	// Bob can't start a group on Alice's food (that edits her catalog row), and nothing is created.
	do(t, api, h.Bob, http.MethodPost, "/diet/foods", map[string]any{
		"name": "Ground beef 80/20", "serving_type": "g", "serving_amount": 100, "calories": 250, "related_food_id": lean,
	}, http.StatusForbidden)
	if foods, _ := foodNames(t, api, h.Bob); contains(foods, "Ground beef 80/20") {
		t.Fatalf("a refused food was created anyway: %v", foods)
	}

	// Alice can, and once the group exists anyone can join it without touching her row.
	food(t, api, h.Alice, "Ground beef 85/15", map[string]any{"related_food_id": lean})
	var before, after struct {
		VariantGroupID *uint
		UpdatedAt      time.Time
	}
	h.Owner.Raw("SELECT variant_group_id, updated_at FROM foods WHERE id = ?", lean).Scan(&before)
	bobs := food(t, api, h.Bob, "Ground beef 80/20", map[string]any{"related_food_id": lean})
	h.Owner.Raw("SELECT variant_group_id, updated_at FROM foods WHERE id = ?", lean).Scan(&after)
	if before.VariantGroupID == nil || after.VariantGroupID == nil || *after.VariantGroupID != *before.VariantGroupID || !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatalf("Alice's food changed when Bob joined its group: %+v -> %+v", before, after)
	}
	var group *uint
	h.Owner.Raw("SELECT variant_group_id FROM foods WHERE id = ?", bobs).Scan(&group)
	if group == nil || *group != *before.VariantGroupID {
		t.Fatalf("Bob's food group %v, want %v", group, before.VariantGroupID)
	}

	// An unknown related food is a 400.
	do(t, api, h.Bob, http.MethodPost, "/diet/foods", map[string]any{
		"name": "Nothing", "serving_type": "g", "serving_amount": 1, "calories": 1, "related_food_id": 999999,
	}, http.StatusBadRequest)
}

func TestDiet_daysAndPlansArePerUser(t *testing.T) {
	h, api := dietAPI(t)
	aliceDay, alicePlan := today(t, api, h.Alice)

	// Bob has nothing of Alice's, and reading today creates his own day and plan (same date: no collision).
	bobDay, bobPlan := today(t, api, h.Bob)
	if bobDay == aliceDay || bobPlan == alicePlan {
		t.Fatalf("Bob got Alice's day %d/%d or plan %d/%d", bobDay, aliceDay, bobPlan, alicePlan)
	}
	missing := do(t, api, h.Bob, http.MethodGet, "/diet/logs/day/999999", nil, http.StatusNotFound)
	other := do(t, api, h.Bob, http.MethodGet, fmt.Sprintf("/diet/logs/day/%d", aliceDay), nil, http.StatusNotFound)
	if string(missing.Body) != string(other.Body) {
		t.Fatalf("Alice's day answers %s, a missing day %s: they must match", other.Body, missing.Body)
	}
	do(t, api, h.Alice, http.MethodGet, fmt.Sprintf("/diet/logs/day/%d", aliceDay), nil, http.StatusOK)

	var week struct {
		Days []idJSON `json:"days"`
	}
	do(t, api, h.Bob, http.MethodGet, "/diet/logs/week", nil, http.StatusOK).JSON(t, &week)
	if len(week.Days) != 1 || week.Days[0].ID != bobDay {
		t.Fatalf("Bob's week %+v, want only his day %d", week.Days, bobDay)
	}

	// Plans: each user's own list; editing Alice's plan as Bob is the missing-ID 404.
	var plans struct {
		Plans []idJSON `json:"plans"`
	}
	do(t, api, h.Bob, http.MethodGet, "/diet/plans/plan/all", nil, http.StatusOK).JSON(t, &plans)
	if len(plans.Plans) != 1 || plans.Plans[0].ID != bobPlan {
		t.Fatalf("Bob's plans %+v, want only %d", plans.Plans, bobPlan)
	}
	macros := map[string]any{"calories": 1800, "protein": 120, "fiber": 25, "carbs": 150, "fat": 60}
	do(t, api, h.Bob, http.MethodPut, fmt.Sprintf("/diet/plans/plan/%d", alicePlan), macros, http.StatusNotFound)
	do(t, api, h.Bob, http.MethodPut, "/diet/plans/plan/999999", macros, http.StatusNotFound)
	do(t, api, h.Alice, http.MethodPut, fmt.Sprintf("/diet/plans/plan/%d", alicePlan), macros, http.StatusOK)
	var goals struct {
		Calories float64 `json:"calories"`
	}
	do(t, api, h.Bob, http.MethodGet, "/diet/logs/goals/today", nil, http.StatusOK).JSON(t, &goals)
	if goals.Calories == 1800 {
		t.Fatal("Alice's new targets showed up as Bob's goals")
	}
}

func TestDiet_mealsAreInvisibleToOthers(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Alice, "Rice", nil)
	aliceDay, _ := today(t, api, h.Alice)
	today(t, api, h.Bob)

	// Alice logs a meal on her day.
	var logged struct {
		Day struct {
			Logs []struct {
				MealID uint `json:"meal_id"`
			} `json:"loggedMeals"`
		} `json:"day"`
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/meal/logedited", map[string]any{
		"meal": mealBody("Lunch", rice, nil), "day_id": aliceDay,
	}, http.StatusOK).JSON(t, &logged)
	if len(logged.Day.Logs) != 1 {
		t.Fatalf("Alice's logs %+v", logged.Day.Logs)
	}
	aliceMeal := logged.Day.Logs[0].MealID

	// Reads: same 404 as a missing meal; lists are empty.
	missing := do(t, api, h.Bob, http.MethodGet, "/diet/meals/meal/999999", nil, http.StatusNotFound)
	other := do(t, api, h.Bob, http.MethodGet, fmt.Sprintf("/diet/meals/meal/%d", aliceMeal), nil, http.StatusNotFound)
	if string(missing.Body) != string(other.Body) {
		t.Fatalf("Alice's meal answers %s, a missing meal %s", other.Body, missing.Body)
	}
	do(t, api, h.Alice, http.MethodGet, fmt.Sprintf("/diet/meals/meal/%d", aliceMeal), nil, http.StatusOK)
	var meals struct {
		Meals []idJSON `json:"meals"`
	}
	do(t, api, h.Bob, http.MethodGet, "/diet/meals/meal/all", nil, http.StatusOK).JSON(t, &meals)
	if len(meals.Meals) != 0 {
		t.Fatalf("Bob sees meals %+v", meals.Meals)
	}

	// Writes against Alice's day or meal are the missing-ID 404, and change nothing.
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/logedited", map[string]any{
		"meal": mealBody("Intruder", rice, nil), "day_id": aliceDay,
	}, http.StatusNotFound)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/editlogged", map[string]any{
		"meal": mealBody("Intruder", rice, nil), "oldMealID": aliceMeal, "day_id": aliceDay,
	}, http.StatusNotFound)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/log-planned", map[string]any{"meal_id": aliceMeal, "day_id": aliceDay}, http.StatusNotFound)
	do(t, api, h.Bob, http.MethodDelete, "/diet/meals/meal/logged", map[string]any{"meal_id": aliceMeal, "day_id": aliceDay}, http.StatusNotFound)
	// Bob's own day with Alice's meal ID: nothing to unlog or edit there.
	bobDay, _ := today(t, api, h.Bob)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/editlogged", map[string]any{
		"meal": mealBody("Intruder", rice, nil), "oldMealID": aliceMeal, "day_id": bobDay,
	}, http.StatusNotFound)

	var n int64
	h.Owner.Raw("SELECT count(*) FROM meals WHERE name = 'Intruder'").Scan(&n)
	if n != 0 {
		t.Fatalf("%d meals created by refused requests", n)
	}
	var items int64
	h.Owner.Raw("SELECT count(*) FROM meal_items WHERE meal_id = ?", aliceMeal).Scan(&items)
	if items != 1 {
		t.Fatalf("Alice's meal has %d items, want 1", items)
	}

	// Bob can't use her private rows as ingredients (a quick-entry food or a recipe): a 400.
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/quick-log", map[string]any{"name": "Secret snack", "calories": 400}, http.StatusOK)
	var quickFood uint
	h.Owner.Raw("SELECT id FROM foods WHERE quick_entry").Scan(&quickFood)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/new", map[string]any{"meal": mealBody("X", quickFood, nil)}, http.StatusBadRequest)
	var recipe struct {
		CF idJSON `json:"composite_food"`
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/composite-food/new", map[string]any{
		"name": "Pico", "items": []map[string]any{{"food_id": rice, "amount": 1}},
	}, http.StatusCreated).JSON(t, &recipe)
	bobFood := food(t, api, h.Bob, "Bread", nil)
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/new", map[string]any{"meal": mealBody("X", bobFood, map[string]any{
		"items": []map[string]any{{"food_id": bobFood, "amount": 1, "composite_food_id": recipe.CF.ID}},
	})}, http.StatusBadRequest)
}

func TestDiet_forgedOwnerIsIgnored(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Bob, "Rice", nil)
	var out struct {
		MealID uint `json:"meal_id"`
	}
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/meal/new", map[string]any{"meal": mealBody("Mine", rice, map[string]any{
		"user_id": h.Alice.ID, "UserID": h.Alice.ID, "ID": 1, "deleted_at": "2020-01-01T00:00:00Z",
	}), "user_id": h.Alice.ID}, http.StatusCreated).JSON(t, &out)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM meals WHERE id = ?", out.MealID).Scan(&owner).Error; err != nil || owner != h.Bob.ID {
		t.Fatalf("meal owner %d (%v), want Bob %d", owner, err, h.Bob.ID)
	}
	h.Owner.Raw("SELECT user_id FROM meal_items WHERE meal_id = ?", out.MealID).Scan(&owner)
	if owner != h.Bob.ID {
		t.Fatalf("meal item owner %d, want Bob", owner)
	}
}

func TestDiet_savedMealsAndPlannedMeals(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Alice, "Rice", nil)
	today(t, api, h.Alice)
	today(t, api, h.Bob)

	var created struct {
		ID uint `json:"saved_meal_id"`
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{
		"name": "Bowl", "items": []map[string]any{{"food_id": rice, "amount": 2}},
	}, http.StatusCreated).JSON(t, &created)
	saved := created.ID
	path := fmt.Sprintf("/diet/meals/saved-meal/%d", saved)

	do(t, api, h.Alice, http.MethodGet, path, nil, http.StatusOK)
	missing := do(t, api, h.Bob, http.MethodGet, "/diet/meals/saved-meal/999999", nil, http.StatusNotFound)
	for _, req := range []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, path, nil},
		{http.MethodPut, path, map[string]any{"name": "Hijacked", "items": []map[string]any{{"food_id": rice, "amount": 9}}}},
		{http.MethodDelete, path + "?force=true", nil},
		{http.MethodDelete, path, nil},
	} {
		res := do(t, api, h.Bob, req.method, req.path, req.body, http.StatusNotFound)
		if string(res.Body) != string(missing.Body) {
			t.Errorf("%s %s: %s, want the missing-ID answer %s", req.method, req.path, res.Body, missing.Body)
		}
	}
	var list struct {
		Saved []idJSON `json:"saved_meals"`
	}
	do(t, api, h.Bob, http.MethodGet, "/diet/meals/saved-meal/all", nil, http.StatusOK).JSON(t, &list)
	if len(list.Saved) != 0 {
		t.Fatalf("Bob sees saved meals %+v", list.Saved)
	}
	var name string
	h.Owner.Raw("SELECT name FROM saved_meals WHERE id = ?", saved).Scan(&name)
	if name != "Bowl" {
		t.Fatalf("Alice's saved meal is now %q", name)
	}

	// Planning from someone else's saved meal is the same 404; Alice plans, copies and reorders her own.
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": saved}, http.StatusNotFound)
	for i := 0; i < 2; i++ {
		do(t, api, h.Alice, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": saved}, http.StatusOK)
	}
	var day struct {
		Day struct {
			Planned []idJSON `json:"plannedMeals"`
		} `json:"day"`
	}
	do(t, api, h.Alice, http.MethodGet, "/diet/logs/today", nil, http.StatusOK).JSON(t, &day)
	if len(day.Day.Planned) != 2 {
		t.Fatalf("Alice planned %d meals, want 2", len(day.Day.Planned))
	}
	ids := []uint{day.Day.Planned[1].ID, day.Day.Planned[0].ID}

	do(t, api, h.Bob, http.MethodPost, "/diet/meals/planned/reorder", map[string]any{"planned_meal_ids": ids}, http.StatusBadRequest)
	do(t, api, h.Bob, http.MethodDelete, "/diet/meals/planned", map[string]any{"planned_meal_id": ids[0]}, http.StatusNotFound)
	copied := do(t, api, h.Bob, http.MethodPost, "/diet/meals/planned/copy-day", map[string]any{"from_offset": 0, "to_offsets": []int{1}}, http.StatusOK)
	var copy struct {
		N int `json:"copied_meals"`
	}
	copied.JSON(t, &copy)
	if copy.N != 0 {
		t.Fatalf("Bob copied %d of Alice's planned meals", copy.N)
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/planned/reorder", map[string]any{"planned_meal_ids": ids}, http.StatusOK)
	do(t, api, h.Alice, http.MethodDelete, "/diet/meals/planned", map[string]any{"planned_meal_id": ids[0]}, http.StatusOK)

	// Alice can still replace and delete her own.
	do(t, api, h.Alice, http.MethodPut, path, map[string]any{"name": "Bowl 2", "items": []map[string]any{{"food_id": rice, "amount": 1}}}, http.StatusOK)
	do(t, api, h.Alice, http.MethodDelete, path+"?force=true", nil, http.StatusNoContent)
}

func TestDiet_recipesArePerUser(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Alice, "Rice", nil)
	body := map[string]any{"name": "Pico", "items": []map[string]any{{"food_id": rice, "amount": 1}}}

	do(t, api, h.Alice, http.MethodPost, "/diet/meals/composite-food/new", body, http.StatusCreated)
	if _, c := foodNames(t, api, h.Alice); !contains(c, "Pico") {
		t.Fatalf("Alice's recipes %v", c)
	}
	if _, c := foodNames(t, api, h.Bob); len(c) != 0 {
		t.Fatalf("Bob sees recipes %v", c)
	}
	// The same name is free for Bob, but not twice for Alice.
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/composite-food/new", body, http.StatusCreated)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/composite-food/new", body, http.StatusConflict)
	// An unknown ingredient is a 400.
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/composite-food/new",
		map[string]any{"name": "Ghost", "items": []map[string]any{{"food_id": 999999, "amount": 1}}}, http.StatusBadRequest)
}

// The same client_id from two users is two rows, each user's replay stays idempotent, and nobody
// waits on the other's advisory lock.
func TestDiet_clientIDsArePerUser(t *testing.T) {
	h, api := dietAPI(t)
	cid := "6f1c2c1e-0000-4000-8000-0000000000aa"
	quick := func(u twousers.User) {
		do(t, api, u, http.MethodPost, "/diet/meals/quick-log", map[string]any{"name": "Bar", "calories": 200, "client_id": cid}, http.StatusOK)
	}
	quick(h.Alice)
	quick(h.Bob)
	quick(h.Alice) // replay
	quick(h.Bob)   // replay
	for _, u := range []twousers.User{h.Alice, h.Bob} {
		var n int64
		h.Owner.Raw("SELECT count(*) FROM meals WHERE client_id = ? AND user_id = ?", cid, u.ID).Scan(&n)
		if n != 1 {
			t.Errorf("%s has %d meals for the client_id, want 1", u.Username, n)
		}
	}

	// Planned meals too.
	rice := food(t, api, h.Alice, "Rice", nil)
	var created struct {
		ID uint `json:"saved_meal_id"`
	}
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{"name": "Bowl", "items": []map[string]any{{"food_id": rice, "amount": 1}}}, http.StatusCreated).JSON(t, &created)
	bobRice := food(t, api, h.Bob, "Bread", nil)
	var bobs struct {
		ID uint `json:"saved_meal_id"`
	}
	do(t, api, h.Bob, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{"name": "Toast", "items": []map[string]any{{"food_id": bobRice, "amount": 1}}}, http.StatusCreated).JSON(t, &bobs)
	pcid := "6f1c2c1e-0000-4000-8000-0000000000bb"
	for i := 0; i < 2; i++ {
		do(t, api, h.Alice, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": created.ID, "client_id": pcid}, http.StatusOK)
		do(t, api, h.Bob, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": bobs.ID, "client_id": pcid}, http.StatusOK)
	}
	for _, u := range []twousers.User{h.Alice, h.Bob} {
		var n int64
		h.Owner.Raw("SELECT count(*) FROM planned_meals WHERE client_id = ? AND user_id = ?", pcid, u.ID).Scan(&n)
		if n != 1 {
			t.Errorf("%s has %d planned meals for the client_id, want 1", u.Username, n)
		}
	}
}

func TestDiet_lockIsPerUser(t *testing.T) {
	h, _ := dietAPI(t)
	cid := "6f1c2c1e-0000-4000-8000-0000000000cc"

	// Alice's request is mid-flight holding the lock for the client_id.
	aliceCtx, endAlice := h.Ctx(h.Alice)
	if err := database.XactLock(aliceCtx, database.From(aliceCtx), "meals.client_id", cid); err != nil {
		t.Fatal(err)
	}
	defer endAlice(false)

	done := make(chan error, 1)
	go func() {
		done <- h.Run(h.Bob, func(ctx context.Context) error {
			_, err := dietrepo.QuickLogMeal(ctx, dietrepo.QuickLogParams{
				DisplayName: "Bar", FoodRowName: "Bar [ql-1]", Calories: 100, ClientID: cid,
			})
			return err
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Bob's quick log waits on Alice's advisory lock for the same client_id")
	}
}

// Row-level security alone, with no scoping in the query: as Bob, unscoped SQL sees and changes
// nothing of Alice's, and the schema refuses rows that point across users.
func TestDiet_rlsBackstop(t *testing.T) {
	h, api := dietAPI(t)
	rice := food(t, api, h.Alice, "Rice", nil)
	day, plan := today(t, api, h.Alice)
	meal := newMeal(t, api, h.Alice, "Lunch", rice)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/meal/logedited", map[string]any{"meal": mealBody("Dinner", rice, nil), "day_id": day}, http.StatusOK)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/saved-meal/new", map[string]any{"name": "Bowl", "items": []map[string]any{{"food_id": rice, "amount": 1}}}, http.StatusCreated)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/composite-food/new", map[string]any{"name": "Pico", "items": []map[string]any{{"food_id": rice, "amount": 1}}}, http.StatusCreated)
	var saved uint
	h.Owner.Raw("SELECT id FROM saved_meals").Scan(&saved)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/planned/from-saved", map[string]any{"saved_meal_id": saved}, http.StatusOK)
	do(t, api, h.Alice, http.MethodPost, "/diet/meals/quick-log", map[string]any{"name": "Snack", "calories": 300}, http.StatusOK)

	owned := []string{"plans", "days", "meals", "planned_meals", "saved_meals", "composite_foods",
		"day_logs", "meal_items", "saved_meal_items", "composite_food_items"}
	for _, tb := range owned {
		var total int64
		h.Owner.Raw("SELECT count(*) FROM " + tb).Scan(&total)
		if total == 0 {
			t.Fatalf("%s: arrange step left no rows for Alice", tb)
		}
		err := h.Run(h.Bob, func(ctx context.Context) error {
			db := database.From(ctx)
			var seen int64
			if err := db.Raw("SELECT count(*) FROM " + tb).Scan(&seen).Error; err != nil {
				return err
			}
			if seen != 0 {
				t.Errorf("%s: Bob sees %d of Alice's %d rows", tb, seen, total)
			}
			if res := db.Exec("UPDATE " + tb + " SET updated_at = now()"); res.Error != nil || res.RowsAffected != 0 {
				t.Errorf("%s: Bob's update touched %d rows (%v)", tb, res.RowsAffected, res.Error)
			}
			if res := db.Exec("DELETE FROM " + tb); res.Error != nil || res.RowsAffected != 0 {
				t.Errorf("%s: Bob's delete touched %d rows (%v)", tb, res.RowsAffected, res.Error)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// A row for Alice written as Bob is refused outright.
	err := h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO plans (name, user_id) VALUES ('x', ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Error("Bob inserted a plan for Alice")
	}

	// Foods: catalog rows are readable, but only the creator or an admin changes them, and only an
	// admin deletes; a quick-entry food is visible to its creator alone (not even to the admin).
	err = h.Run(h.Bob, func(ctx context.Context) error {
		db := database.From(ctx)
		var seen int64
		db.Raw("SELECT count(*) FROM foods WHERE id = ?", rice).Scan(&seen)
		if seen != 1 {
			t.Errorf("Bob can't read the catalog food (%d)", seen)
		}
		db.Raw("SELECT count(*) FROM foods WHERE quick_entry").Scan(&seen)
		if seen != 0 {
			t.Errorf("Bob sees %d of Alice's quick-entry foods", seen)
		}
		if res := db.Exec("UPDATE foods SET calories = 1 WHERE id = ?", rice); res.RowsAffected != 0 {
			t.Errorf("Bob edited Alice's catalog food (%d rows)", res.RowsAffected)
		}
		if res := db.Exec("UPDATE foods SET deleted_at = now() WHERE id = ?", rice); res.Error == nil && res.RowsAffected != 0 {
			t.Errorf("Bob soft-deleted Alice's food (%d rows)", res.RowsAffected)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO foods (name, serving_type, serving_amount, calories, created_by) VALUES ('Forged', 'g', 1, 1, ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Error("Bob added a catalog food as Alice")
	}
	_ = h.Run(h.Admin, func(ctx context.Context) error {
		var seen int64
		database.From(ctx).Raw("SELECT count(*) FROM foods WHERE quick_entry").Scan(&seen)
		if seen != 0 {
			t.Errorf("the admin sees %d of Alice's quick-entry foods", seen)
		}
		return nil
	})

	// FK checks bypass RLS, so the composite FKs are what stop a row pointing at someone else's
	// parent. As the schema owner (no RLS) each of these must still be refused.
	for name, stmt := range map[string]string{
		"meal item on Alice's meal":   fmt.Sprintf("INSERT INTO meal_items (meal_id, food_id, amount, user_id) VALUES (%d, %d, 1, %d)", meal, rice, h.Bob.ID),
		"day log on Alice's day":      fmt.Sprintf("INSERT INTO day_logs (day_id, meal_id, user_id) VALUES (%d, %d, %d)", day, meal, h.Bob.ID),
		"day on Alice's plan":         fmt.Sprintf("INSERT INTO days (date, plan_id, user_id) VALUES ('2030-01-01', %d, %d)", plan, h.Bob.ID),
		"planned meal from her saved": fmt.Sprintf("INSERT INTO planned_meals (day_id, meal_id, saved_meal_id, user_id) VALUES (%d, %d, %d, %d)", day, meal, saved, h.Bob.ID),
		"saved meal item on hers":     fmt.Sprintf("INSERT INTO saved_meal_items (saved_meal_id, food_id, amount, user_id) VALUES (%d, %d, 1, %d)", saved, rice, h.Bob.ID),
	} {
		if err := h.Owner.Exec(stmt).Error; err == nil {
			t.Errorf("%s: the schema accepted it", name)
		}
	}
}
