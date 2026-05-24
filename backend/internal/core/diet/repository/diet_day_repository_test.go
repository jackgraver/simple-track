package dietrepo_test

import (
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/testutil"
	"context"
	"strings"
	"testing"
	"time"

	"be-simpletracker/internal/utils"
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
	testutil.SetupTestDB(t)
	_, err := dietrepo.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName:   "X",
		FoodRowName:   "X [ql-1]",
		ReplaceMealID: 9999,
	})
	if err == nil || !strings.Contains(err.Error(), "replace_meal_id") {
		t.Fatalf("expected replace error, got %v", err)
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
