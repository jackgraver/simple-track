package dietrepo_test

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/testutil"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"be-simpletracker/internal/utils"
	"gorm.io/gorm"
)

func TestDayMealPlanToday_createsDayWithDefaultPlan(t *testing.T) {
	testutil.SetupTestDB(t)
	day, err := dietrepo.DayMealPlanToday(0)
	if err != nil {
		t.Fatal(err)
	}
	if day.ID == 0 || day.Plan.ID == 0 {
		t.Fatalf("day %+v", day)
	}
}

func TestDayByID_andCalculateTotals(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "Bulk")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Steak", models.Food{
		ServingType: "g", ServingAmount: 100,
		Calories: 250, Protein: 26, Fiber: 0, Carbs: 0, Fat: 15,
	})
	meal := testutil.SeedMeal(t, db, "Dinner", food.ID, 2)
	testutil.SeedDayLog(t, db, day.ID, meal.ID)

	loaded, err := dietrepo.DayByID(int(day.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Logs) != 1 {
		t.Fatalf("logs %+v", loaded.Logs)
	}
	tot := dietrepo.CalculateTotals(day.ID)
	if tot.Calories != 500 || tot.Protein != 52 {
		t.Fatalf("totals %+v", tot)
	}
}

func TestGoalsToday_returnsPlanForToday(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "Maintain")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	got, err := dietrepo.GoalsToday()
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Maintain" {
		t.Fatalf("got %+v", got)
	}
}

func TestFindDayByDate_createsIfMissing(t *testing.T) {
	testutil.SetupTestDB(t)
	day, err := dietrepo.FindDayByDate(testutil.Today())
	if err != nil || day == nil || day.ID == 0 {
		t.Fatalf("day %+v err %v", day, err)
	}
}

func TestDaysByDateRange_andDayByIDGeneric(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	today := testutil.Today()
	testutil.SeedDay(t, db, today, plan.ID)
	testutil.SeedDay(t, db, today.AddDate(0, 0, -1), plan.ID)
	ctx := context.Background()
	days, err := dietrepo.DaysByDateRange(ctx, today.AddDate(0, 0, -2), today)
	if err != nil || len(days) < 2 {
		t.Fatalf("days %+v err %v", days, err)
	}
	var firstID uint
	if err := db.Model(&models.DietDay{}).Order("id ASC").Limit(1).Pluck("id", &firstID).Error; err != nil {
		t.Fatal(err)
	}
	generic, err := dietrepo.DayByIDGeneric(ctx, firstID)
	if err != nil || generic.ID != firstID {
		t.Fatalf("generic %+v err %v", generic, err)
	}
}

func TestCountUnloggedPlannedMealsPerCalendarDay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	today := testutil.Today()
	day := testutil.SeedDay(t, db, today, plan.ID)
	food := testutil.SeedFood(t, db, "Apple", testutil.DefaultMacros())
	m1 := testutil.SeedMeal(t, db, "Lunch", food.ID, 1)
	m2 := testutil.SeedMeal(t, db, "Snack", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, day.ID, m1.ID, 0, false)
	testutil.SeedPlannedMeal(t, db, day.ID, m2.ID, 1, true)
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start
	counts, err := dietrepo.CountUnloggedPlannedMealsPerCalendarDay(start, end)
	if err != nil {
		t.Fatal(err)
	}
	if counts[start.Format("2006-01-02")] != 1 {
		t.Fatalf("counts %+v", counts)
	}
}

func TestAllMealDays(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	days, err := dietrepo.AllMealDays()
	if err != nil || len(days) != 1 {
		t.Fatalf("days %+v err %v", days, err)
	}
}

func TestCreateDayMeal_andDayLogExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Banana", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "B", food.ID, 1)
	if err := dietrepo.CreateDayMeal(&models.DayLog{DayID: day.ID, MealID: meal.ID}); err != nil {
		t.Fatal(err)
	}
	ok, err := dietrepo.DayLogExists(day.ID, meal.ID)
	if err != nil || !ok {
		t.Fatalf("exists %v err %v", ok, err)
	}
}

func TestDeleteLoggedMeal_andUpdateDayLogMeal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Tofu", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old", food.ID, 1)
	newMeal := testutil.SeedMeal(t, db, "New", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	if err := dietrepo.UpdateDayLogMeal(day.ID, oldMeal.ID, newMeal.ID); err != nil {
		t.Fatal(err)
	}
	ok, _ := dietrepo.DayLogExists(day.ID, newMeal.ID)
	if !ok {
		t.Fatal("expected new meal logged")
	}
	if err := dietrepo.DeleteLoggedMeal(day.ID, newMeal.ID); err != nil {
		t.Fatal(err)
	}
	ok, _ = dietrepo.DayLogExists(day.ID, newMeal.ID)
	if ok {
		t.Fatal("expected deleted")
	}
}

func TestSetPlannedMealLogged(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Pasta", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Dinner", food.ID, 1)
	pm := testutil.SeedPlannedMeal(t, db, day.ID, meal.ID, 0, false)
	if err := dietrepo.SetPlannedMealLogged(day.ID, meal.ID); err != nil {
		t.Fatal(err)
	}
	var reloaded models.PlannedMeal
	if err := db.First(&reloaded, pm.ID).Error; err != nil || !reloaded.Logged {
		t.Fatalf("reloaded %+v", reloaded)
	}
}

func TestQuickLogMeal_createsLog(t *testing.T) {
	testutil.SetupTestDB(t)
	dayID, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName: "Protein Shake",
		FoodRowName: "Protein Shake [ql-99]",
		Calories:    200,
		Protein:     30,
	})
	if err != nil || dayID == 0 {
		t.Fatalf("dayID %d err %v", dayID, err)
	}
	tot := dietrepo.CalculateTotals(dayID)
	if tot.Calories != 200 {
		t.Fatalf("totals %+v", tot)
	}
}

func TestQuickLogMeal_replaceExistingLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Bar", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old Bar", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	_, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName:   "New Bar",
		FoodRowName:   "New Bar [ql-1]",
		Calories:      150,
		ReplaceMealID: oldMeal.ID,
		Offset:        0,
	})
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := dietrepo.DayLogExists(day.ID, oldMeal.ID)
	if ok {
		t.Fatal("old meal should be replaced")
	}
}

func TestQuickLogMeal_replaceInvalid(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName:   "X",
		FoodRowName:   "X [ql-1]",
		ReplaceMealID: 9999,
	})
	if err == nil || !strings.Contains(err.Error(), "replace_meal_id") {
		t.Fatalf("expected replace error, got %v", err)
	}
	// The edit is one DB transaction: the failed replace leaves no new food, meal or log behind.
	for _, m := range []any{&models.Food{}, &models.Meal{}, &models.DayLog{}} {
		var n int64
		db.Unscoped().Model(m).Count(&n)
		if n != 0 {
			t.Fatalf("%T rows after failed replace = %d, want 0", m, n)
		}
	}
}

// A failed replace keeps the original log: the edit cannot lose the meal.
func TestQuickLogMeal_replaceFailureKeepsOriginal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Bar", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old Bar", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	// Wrong day (tomorrow): the replace target is not logged there, so the whole call rolls back.
	if _, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName:   "New Bar",
		FoodRowName:   "New Bar [ql-2]",
		ReplaceMealID: oldMeal.ID,
		Offset:        1,
	}); err == nil {
		t.Fatal("expected replace error")
	}
	if ok, _ := dietrepo.DayLogExists(day.ID, oldMeal.ID); !ok {
		t.Fatal("original log should survive a failed replace")
	}
	var meals int64
	db.Model(&models.Meal{}).Count(&meals)
	if meals != 1 {
		t.Fatalf("meals = %d, want 1 (no duplicate)", meals)
	}
}

func TestEditLoggedMeal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Fish", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old Fish", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	newMeal := &models.Meal{
		Name: "New Fish",
		Items: []models.MealItem{{
			FoodID: food.ID,
			Amount: 2,
		}},
	}
	if err := dietrepo.EditLoggedMeal(day.ID, oldMeal.ID, newMeal); err != nil {
		t.Fatal(err)
	}
	ok, _ := dietrepo.DayLogExists(day.ID, oldMeal.ID)
	if ok {
		t.Fatal("old meal should be unlinked")
	}
}

func TestPlansGetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SeedPlan(t, db, "A")
	testutil.SeedPlan(t, db, "B")
	ctx := context.Background()
	all, err := dietrepo.PlansGetAll(ctx, utils.QueryParams{})
	if err != nil || len(all.Data) != 2 {
		t.Fatalf("all %+v err %v", all, err)
	}
	paged, err := dietrepo.PlansGetAll(ctx, utils.QueryParams{Page: 1, PageSize: 1})
	if err != nil || paged.Pagination == nil || len(paged.Data) != 1 {
		t.Fatalf("paged %+v err %v", paged, err)
	}
}

func TestDayByIDGeneric_notFound(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := dietrepo.DayByIDGeneric(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCopyPlannedMeals_appendsInOrderWithOwnMeals(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	src := testutil.SeedDay(t, db, clock.ZerodTime(0), plan.ID)
	dst := testutil.SeedDay(t, db, clock.ZerodTime(-1), plan.ID)
	food := testutil.SeedFood(t, db, "Oats", models.Food{ServingAmount: 1, Calories: 100})
	a := testutil.SeedMeal(t, db, "A", food.ID, 1)
	b := testutil.SeedMeal(t, db, "B", food.ID, 2)
	testutil.SeedPlannedMeal(t, db, src.ID, b.ID, 1, false)
	testutil.SeedPlannedMeal(t, db, src.ID, a.ID, 0, false)
	logged := testutil.SeedMeal(t, db, "Eaten", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, src.ID, logged.ID, 2, true)
	existing := testutil.SeedMeal(t, db, "Existing", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, dst.ID, existing.ID, 0, false)

	n, err := dietrepo.CopyPlannedMeals(0, []int{-1, 0, -1})
	if err != nil || n != 2 {
		t.Fatalf("n %d err %v", n, err)
	}
	var got []models.PlannedMeal
	db.Where("day_id = ?", dst.ID).Order("display_order").Preload("Meal.Items").Find(&got)
	names := []string{}
	for _, p := range got {
		names = append(names, p.Meal.Name)
		if p.Meal.Name != "Existing" && (p.MealID == a.ID || p.MealID == b.ID) {
			t.Fatalf("copy shares the source meal row")
		}
	}
	if strings.Join(names, ",") != "Existing,A,B" {
		t.Fatalf("got %v", names)
	}
	if got[2].Meal.Items[0].Amount != 2 {
		t.Fatalf("items not copied: %+v", got[2].Meal.Items)
	}
	var srcCount int64
	db.Model(&models.PlannedMeal{}).Where("day_id = ?", src.ID).Count(&srcCount)
	if srcCount != 3 {
		t.Fatalf("source changed: %d", srcCount)
	}
}

func TestLogPlannedMeal_pastDayIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	yesterday := testutil.SeedDay(t, db, testutil.Today().AddDate(0, 0, -1), plan.ID)
	food := testutil.SeedFood(t, db, "Rice", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Lunch", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, yesterday.ID, meal.ID, 0, false)

	for i := 0; i < 2; i++ { // second call is a replay
		if err := dietrepo.LogPlannedMeal(yesterday.ID, meal.ID); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	var n int64
	db.Model(&models.DayLog{}).Where("day_id = ? AND meal_id = ?", yesterday.ID, meal.ID).Count(&n)
	if n != 1 {
		t.Fatalf("want 1 day log, got %d", n)
	}
	var pm models.PlannedMeal
	db.Where("day_id = ?", yesterday.ID).First(&pm)
	if !pm.Logged {
		t.Fatal("planned meal should be marked logged")
	}
	if err := dietrepo.LogPlannedMeal(yesterday.ID, meal.ID+999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("want ErrRecordNotFound, got %v", err)
	}
}

func TestDeleteLoggedMeal_idempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Egg", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Eggs", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, meal.ID)
	for i := 0; i < 2; i++ {
		if err := dietrepo.DeleteLoggedMeal(day.ID, meal.ID); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
}

func TestQuickLogMeal_clientIDIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	cid := "11111111-1111-4111-8111-111111111111"
	for i := 0; i < 2; i++ {
		if _, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
			DisplayName: "Bar", FoodRowName: fmt.Sprintf("Bar [ql-%d]", i), Calories: 200, ClientID: cid,
		}); err != nil {
			t.Fatal(err)
		}
	}
	var n int64
	db.Model(&models.DayLog{}).Count(&n)
	if n != 1 {
		t.Fatalf("want 1 day log after replay, got %d", n)
	}
}

// A replay of from-saved after its planned meal was deleted is still a replay (the unique index
// on client_id covers soft-deleted rows).
func TestAddPlannedMealFromSavedMeal_replayAfterDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Oats", testutil.DefaultMacros())
	saved := testutil.SeedSavedMeal(t, db, "Breakfast", food.ID)
	cid := "44444444-4444-4444-8444-444444444444"
	add := func() error {
		meal := models.Meal{Name: saved.Name, Items: []models.MealItem{{FoodID: food.ID, Amount: 1}}}
		return dietrepo.AddPlannedMealFromSavedMeal(0, saved.ID, &meal, cid)
	}
	if err := add(); err != nil {
		t.Fatal(err)
	}
	if err := db.Where("client_id = ?", cid).Delete(&models.PlannedMeal{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := add(); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Unscoped().Model(&models.PlannedMeal{}).Where("client_id = ?", cid).Count(&n)
	if n != 1 {
		t.Fatalf("replay after delete created a new planned meal: %d rows", n)
	}
}
