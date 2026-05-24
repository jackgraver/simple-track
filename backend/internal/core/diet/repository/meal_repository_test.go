package dietrepo_test

import (
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/testutil"
	"strings"
	"testing"
)

func TestMealCreateAndMealByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	food := testutil.SeedFood(t, db, "Yogurt", testutil.DefaultMacros())
	id, err := dietrepo.MealCreate(&models.Meal{
		Name:  "Breakfast",
		Items: []models.MealItem{{FoodID: food.ID, Amount: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	meal, err := dietrepo.MealByID(id)
	if err != nil || meal.Name != "Breakfast" {
		t.Fatalf("meal %+v err %v", meal, err)
	}
}

func TestMealsAll_excludesIDs(t *testing.T) {
	db := testutil.SetupTestDB(t)
	food := testutil.SeedFood(t, db, "Milk", testutil.DefaultMacros())
	a := testutil.SeedMeal(t, db, "A", food.ID, 1)
	testutil.SeedMeal(t, db, "B", food.ID, 1)
	meals, err := dietrepo.MealsAll([]uint{a.ID})
	if err != nil || len(meals) != 1 || meals[0].Name != "B" {
		t.Fatalf("meals %+v err %v", meals, err)
	}
}

func TestSavedMealCRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)
	food := testutil.SeedFood(t, db, "Cheese", testutil.DefaultMacros())
	id, err := dietrepo.SavedMealCreate(&models.SavedMeal{
		Name:  "Snack Plate",
		Items: []models.SavedMealItem{{FoodID: food.ID, Amount: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	sm, err := dietrepo.SavedMealByID(id)
	if err != nil || sm.Name != "Snack Plate" {
		t.Fatalf("sm %+v err %v", sm, err)
	}
	all, err := dietrepo.SavedMealsAll(nil)
	if err != nil || len(all) != 1 {
		t.Fatalf("all %+v err %v", all, err)
	}
	if err := dietrepo.SavedMealReplace(id, &models.SavedMeal{
		Name:  "Updated Plate",
		Items: []models.SavedMealItem{{FoodID: food.ID, Amount: 2}},
	}); err != nil {
		t.Fatal(err)
	}
	updated, err := dietrepo.SavedMealByID(id)
	if err != nil || updated.Name != "Updated Plate" {
		t.Fatalf("updated %+v", updated)
	}
	if err := dietrepo.SavedMealDelete(id); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteSavedMeal_clearsUnloggedPlanned(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Ham", testutil.DefaultMacros())
	sm := testutil.SeedSavedMeal(t, db, "Ham Wrap", food.ID)
	meal := testutil.SeedMeal(t, db, "Wrap", food.ID, 1)
	sid := sm.ID
	pm := models.PlannedMeal{
		DayID: day.ID, MealID: meal.ID, SavedMealID: &sid, Logged: false, DisplayOrder: 0,
	}
	if err := db.Create(&pm).Error; err != nil {
		t.Fatal(err)
	}
	if err := dietrepo.DeleteSavedMeal(sm.ID); err != nil {
		t.Fatal(err)
	}
	count, err := dietrepo.CountUnloggedPlannedBySavedMealID(sm.ID)
	if err != nil || count != 0 {
		t.Fatalf("count %d err %v", count, err)
	}
	var remaining models.SavedMeal
	if err := db.First(&remaining, sm.ID).Error; err == nil {
		t.Fatal("saved meal should be deleted")
	}
}

func TestPlannedMealLifecycle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Salad", testutil.DefaultMacros())
	m1 := testutil.SeedMeal(t, db, "Salad 1", food.ID, 1)
	m2 := testutil.SeedMeal(t, db, "Salad 2", food.ID, 1)
	pm1 := testutil.SeedPlannedMeal(t, db, day.ID, m1.ID, 0, false)
	pm2 := testutil.SeedPlannedMeal(t, db, day.ID, m2.ID, 1, false)
	next, err := dietrepo.NextPlannedMealDisplayOrder(day.ID)
	if err != nil || next != 2 {
		t.Fatalf("next %d err %v", next, err)
	}
	if err := dietrepo.PlannedMealReorder(day.ID, []uint{pm2.ID, pm1.ID}); err != nil {
		t.Fatal(err)
	}
	if err := dietrepo.PlannedMealDelete(pm1.ID, day.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPlannedMealReorder_mismatch(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Soup", testutil.DefaultMacros())
	m1 := testutil.SeedMeal(t, db, "Soup A", food.ID, 1)
	m2 := testutil.SeedMeal(t, db, "Soup B", food.ID, 1)
	pm1 := testutil.SeedPlannedMeal(t, db, day.ID, m1.ID, 0, false)
	pm2 := testutil.SeedPlannedMeal(t, db, day.ID, m2.ID, 1, false)
	err := dietrepo.PlannedMealReorder(day.ID, []uint{pm1.ID, pm1.ID})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	err = dietrepo.PlannedMealReorder(day.ID, []uint{pm1.ID, pm2.ID, pm1.ID})
	if err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestAddPlannedMealFromSavedMeal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Turkey", testutil.DefaultMacros())
	sm := testutil.SeedSavedMeal(t, db, "Turkey Bowl", food.ID)
	meal := &models.Meal{
		Name: sm.Name,
		Items: []models.MealItem{{
			FoodID: food.ID,
			Amount: 1,
		}},
	}
	if err := dietrepo.AddPlannedMealFromSavedMeal(0, sm.ID, meal); err != nil {
		t.Fatal(err)
	}
	count, err := dietrepo.CountUnloggedPlannedBySavedMealID(sm.ID)
	if err != nil || count != 1 {
		t.Fatalf("count %d err %v", count, err)
	}
}

func TestDeleteUnloggedPlannedBySavedMealID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Bean", testutil.DefaultMacros())
	sm := testutil.SeedSavedMeal(t, db, "Bean Bowl", food.ID)
	meal := testutil.SeedMeal(t, db, "Beans", food.ID, 1)
	sid := sm.ID
	if err := db.Create(&models.PlannedMeal{
		DayID: day.ID, MealID: meal.ID, SavedMealID: &sid, Logged: false,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := dietrepo.DeleteUnloggedPlannedBySavedMealID(sm.ID); err != nil {
		t.Fatal(err)
	}
	count, _ := dietrepo.CountUnloggedPlannedBySavedMealID(sm.ID)
	if count != 0 {
		t.Fatalf("count %d", count)
	}
}

func TestFoodCreate(t *testing.T) {
	testutil.SetupTestDB(t)
	f := &models.Food{
		Name: "Direct", ServingType: "g", ServingAmount: 1, Calories: 1,
	}
	if err := dietrepo.FoodCreate(f); err != nil {
		t.Fatal(err)
	}
}
