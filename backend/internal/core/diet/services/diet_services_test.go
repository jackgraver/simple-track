package services_test

import (
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/core/diet/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"testing"
)

func TestMealPlanToday_returnsDayAndTotals(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Rice", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Lunch", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, meal.ID)

	gotDay, tot, err := services.MealPlanToday(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotDay.ID == 0 || tot.Calories != 100 {
		t.Fatalf("day %+v totals %+v", gotDay, tot)
	}
}

func TestMealPlanWeek_andMonth(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	ctx := context.Background()
	week, err := services.MealPlanWeek(ctx)
	if err != nil || len(week) == 0 {
		t.Fatalf("week %+v err %v", week, err)
	}
	days, start, end, month, err := services.MealPlanMonth(ctx, 0)
	if err != nil || len(days) == 0 || month == 0 || !start.Before(end) {
		t.Fatalf("month days=%d start=%v end=%v month=%v err=%v", len(days), start, end, month, err)
	}
}

func TestMonthPlannedSummary(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Carrot", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Side", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, day.ID, meal.ID, 0, false)
	counts, err := services.MonthPlannedSummary(context.Background(), 0)
	if err != nil || len(counts) == 0 {
		t.Fatalf("counts %+v err %v", counts, err)
	}
	found := false
	for _, c := range counts {
		if c > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected planned count in month summary")
	}
}

func TestMealPlanDay_andGoalsToday(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "Goals")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	gotDay, tot, err := services.MealPlanDay(context.Background(), day.ID)
	if err != nil || gotDay.ID != day.ID {
		t.Fatalf("day %+v err %v", gotDay, err)
	}
	if tot.Calories != 0 {
		t.Fatalf("expected zero totals, got %+v", tot)
	}
	goals, err := services.GoalsToday()
	if err != nil || goals.Name != "Goals" {
		t.Fatalf("goals %+v err %v", goals, err)
	}
}

func TestDayWithTotalsHelpers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Nuts", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Snack", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, meal.ID)

	byID, err := services.DayWithTotalsByID(int(day.ID))
	if err != nil || byID.Totals.Calories != 100 {
		t.Fatalf("byID %+v err %v", byID, err)
	}
	loaded, err := services.MealPlanDayByID(int(day.ID))
	if err != nil {
		t.Fatal(err)
	}
	wrap := services.DayWithTotalsForDay(loaded)
	if wrap.Totals.Calories != 100 {
		t.Fatalf("wrap %+v", wrap)
	}
	reloaded, err := services.ReloadDayWithTotals(loaded)
	if err != nil || reloaded.Totals.Calories != 100 {
		t.Fatalf("reloaded %+v err %v", reloaded, err)
	}
}

func TestFindMealPlanDayByRowIDOrToday(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	byID, err := services.FindMealPlanDayByRowIDOrToday(day.ID)
	if err != nil || byID.ID != day.ID {
		t.Fatalf("byID %+v err %v", byID, err)
	}
	today, err := services.FindMealPlanDayByRowIDOrToday(0)
	if err != nil || today.ID == 0 {
		t.Fatalf("today %+v err %v", today, err)
	}
}

func TestFoodAndCompositeServices(t *testing.T) {
	testutil.SetupTestDB(t)
	created, err := services.CreateFood(&models.Food{
		Name: "Broccoli", ServingType: "g", ServingAmount: 100, Calories: 34,
	}, nil)
	if err != nil || created.ID == 0 {
		t.Fatalf("food %+v err %v", created, err)
	}
	all, err := services.AllFoods(nil)
	if err != nil || len(all) != 1 {
		t.Fatalf("all %+v err %v", all, err)
	}
	picker, err := services.AllFoodsForPicker(nil)
	if err != nil || len(picker) != 1 {
		t.Fatalf("picker %+v err %v", picker, err)
	}
	cfID, err := services.CreateCompositeFood(&models.CompositeFood{
		Name: "Mix",
		Items: []models.CompositeFoodItem{{
			FoodID: created.ID,
			Amount: 1,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cf, err := services.CompositeFoodByID(cfID)
	if err != nil || cf.Name != "Mix" {
		t.Fatalf("cf %+v err %v", cf, err)
	}
	composites, err := services.AllCompositeFoods()
	if err != nil || len(composites) != 1 {
		t.Fatalf("composites %+v err %v", composites, err)
	}
}

func TestMealAndLoggingServices(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Potato", testutil.DefaultMacros())
	mealID, err := services.CreateMeal(&models.Meal{
		Name:  "Baked",
		Items: []models.MealItem{{FoodID: food.ID, Amount: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := services.CreateDayMeal(&models.DayLog{DayID: day.ID, MealID: mealID}); err != nil {
		t.Fatal(err)
	}
	ok, err := services.DayLogExistsForMeal(day.ID, mealID)
	if err != nil || !ok {
		t.Fatalf("exists %v err %v", ok, err)
	}
	meal, err := services.MealByID(mealID)
	if err != nil || meal.Name != "Baked" {
		t.Fatalf("meal %+v err %v", meal, err)
	}
	meals, err := services.AllMeals(nil)
	if err != nil || len(meals) != 1 {
		t.Fatalf("meals %+v err %v", meals, err)
	}
	tot := services.CalculateTotals(day.ID)
	if tot.Calories != 100 {
		t.Fatalf("tot %+v", tot)
	}
}

func TestSavedMealServices(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Avocado", testutil.DefaultMacros())
	smID, err := services.CreateSavedMeal(&models.SavedMeal{
		Name:  "Avocado Toast",
		Items: []models.SavedMealItem{{FoodID: food.ID, Amount: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	meal := testutil.SeedMeal(t, db, "Toast", food.ID, 1)
	sid := smID
	if err := db.Create(&models.PlannedMeal{
		DayID: day.ID, MealID: meal.ID, SavedMealID: &sid, Logged: false, DisplayOrder: 0,
	}).Error; err != nil {
		t.Fatal(err)
	}
	info, err := services.SavedMealPlannedUsageInfo(smID)
	if err != nil || info.ReferenceCount != 1 {
		t.Fatalf("info %+v err %v", info, err)
	}
	all, err := services.AllSavedMeals(nil)
	if err != nil || len(all) != 1 {
		t.Fatalf("all %+v err %v", all, err)
	}
	if err := services.ReplaceSavedMeal(smID, &models.SavedMeal{
		Name:  "Updated Toast",
		Items: []models.SavedMealItem{{FoodID: food.ID, Amount: 2}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := services.DeleteSavedMeal(smID); err != nil {
		t.Fatal(err)
	}
}

func TestSavedMealByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	food := testutil.SeedFood(t, db, "Cheddar", testutil.DefaultMacros())
	smID, err := services.CreateSavedMeal(&models.SavedMeal{
		Name:  "Cheese Snack",
		Items: []models.SavedMealItem{{FoodID: food.ID, Amount: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	sm, err := services.SavedMealByID(smID)
	if err != nil || sm.Name != "Cheese Snack" {
		t.Fatalf("sm %+v err %v", sm, err)
	}
}

func TestPlannedMealServices(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Chicken Breast", testutil.DefaultMacros())
	sm := testutil.SeedSavedMeal(t, db, "Meal Prep", food.ID)
	if err := services.AddPlannedMealFromSavedMeal(0, sm.ID); err != nil {
		t.Fatal(err)
	}
	day, err := services.FindMealPlanDay(testutil.Today())
	if err != nil {
		t.Fatal(err)
	}
	if len(day.PlannedMeals) == 0 {
		loaded, err := services.MealPlanDayByID(int(day.ID))
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.PlannedMeals) == 0 {
			t.Fatal("expected planned meal")
		}
		ids := make([]uint, len(loaded.PlannedMeals))
		for i, pm := range loaded.PlannedMeals {
			ids[i] = pm.ID
		}
		if err := services.ReorderPlannedMeals(0, ids); err != nil {
			t.Fatal(err)
		}
		if err := services.DeletePlannedMeal(0, ids[0]); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSetPlannedMealLogged_andDeleteLoggedMeal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Cottage Cheese", testutil.DefaultMacros())
	meal := testutil.SeedMeal(t, db, "Snack", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, day.ID, meal.ID, 0, false)
	testutil.SeedDayLog(t, db, day.ID, meal.ID)
	if err := services.SetPlannedMealLogged(day.ID, meal.ID); err != nil {
		t.Fatal(err)
	}
	if err := services.DeleteLoggedMeal(day.ID, meal.ID); err != nil {
		t.Fatal(err)
	}
}

func TestEditAndQuickLogServices(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Protein Bar", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	if err := services.EditLoggedMeal(day.ID, oldMeal.ID, &models.Meal{
		Name: "Edited",
		Items: []models.MealItem{{FoodID: food.ID, Amount: 2}},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := services.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName: "Shake",
		FoodRowName: "Shake [ql-42]",
		Calories:    180,
		Protein:     25,
	})
	if err != nil || result.Totals.Calories == 0 {
		t.Fatalf("quick log %+v err %v", result, err)
	}
}

func TestUpdateDayLogMeal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Bread", testutil.DefaultMacros())
	oldMeal := testutil.SeedMeal(t, db, "Old", food.ID, 1)
	newMeal := testutil.SeedMeal(t, db, "New", food.ID, 1)
	testutil.SeedDayLog(t, db, day.ID, oldMeal.ID)
	if err := services.UpdateDayLogMeal(day.ID, oldMeal.ID, newMeal.ID); err != nil {
		t.Fatal(err)
	}
	ok, _ := services.DayLogExistsForMeal(day.ID, newMeal.ID)
	if !ok {
		t.Fatal("expected updated log")
	}
}

func TestGetAllPlans_andUpdatePlanMacros(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SeedPlan(t, db, "One")
	testutil.SeedPlan(t, db, "Two")
	all, err := services.GetAllPlans(context.Background(), utils.QueryParams{})
	if err != nil || len(all.Data) != 2 {
		t.Fatalf("all %+v err %v", all, err)
	}
	updated, err := services.UpdatePlanMacros(all.Data[0].ID, 2100, 160, 35, 180, 70)
	if err != nil || updated.Calories != 2100 {
		t.Fatalf("updated %+v err %v", updated, err)
	}
}

func TestAllMealDays_andFindMealPlanDay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	days, err := services.AllMealDays()
	if err != nil || len(days) != 1 {
		t.Fatalf("days %+v err %v", days, err)
	}
	day, err := services.FindMealPlanDay(testutil.Today())
	if err != nil || day.ID == 0 {
		t.Fatalf("day %+v err %v", day, err)
	}
}
