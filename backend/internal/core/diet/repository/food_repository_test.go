package dietrepo_test

import (
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/testutil"
	"testing"
)

func TestEnrichFoodVariants_nilSafe(t *testing.T) {
	testutil.SetupTestDB(t)
	dietrepo.EnrichFoodVariants(nil)
}

func TestFoodCreateAndFoodsAll_excludesQuickEntry(t *testing.T) {
	db := testutil.SetupTestDB(t)
	f := testutil.DefaultMacros()
	f.QuickEntry = false
	testutil.SeedFood(t, db, "Chicken", f)
	q := testutil.DefaultMacros()
	q.QuickEntry = true
	testutil.SeedFood(t, db, "Quick Snack [ql-1]", q)

	foods, err := dietrepo.FoodsAll(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(foods) != 1 || foods[0].Name != "Chicken" {
		t.Fatalf("got %+v", foods)
	}
}

func TestFoodCreateWithOptionalRelated_linksVariantGroup(t *testing.T) {
	db := testutil.SetupTestDB(t)
	base := testutil.SeedFood(t, db, "Rice White", testutil.DefaultMacros())
	related := base.ID
	newFood := &models.Food{
		Name:          "Rice Brown",
		ServingType:   "g",
		ServingAmount: 100,
		Calories:      110,
		Protein:       3,
	}
	if err := dietrepo.FoodCreateWithOptionalRelated(newFood, &related); err != nil {
		t.Fatal(err)
	}
	if newFood.VariantGroupID == nil {
		t.Fatal("expected variant group")
	}
	dietrepo.EnrichFoodVariants(newFood)
	if len(newFood.Variants) != 1 || newFood.Variants[0].ID != base.ID {
		t.Fatalf("variants %+v", newFood.Variants)
	}
}

func TestFoodCreateWithOptionalRelated_noRelated(t *testing.T) {
	testutil.SetupTestDB(t)
	f := &models.Food{
		Name:          "Oats",
		ServingType:   "g",
		ServingAmount: 40,
		Calories:      150,
	}
	if err := dietrepo.FoodCreateWithOptionalRelated(f, nil); err != nil {
		t.Fatal(err)
	}
	if f.ID == 0 {
		t.Fatal("expected id")
	}
}

func TestFoodsAllWithVariantSiblings_andExclude(t *testing.T) {
	db := testutil.SetupTestDB(t)
	a := testutil.SeedFood(t, db, "Bread A", testutil.DefaultMacros())
	gid := uint(7)
	if err := db.Model(&a).Update("variant_group_id", gid).Error; err != nil {
		t.Fatal(err)
	}
	b := testutil.SeedFood(t, db, "Bread B", testutil.DefaultMacros())
	if err := db.Model(&b).Update("variant_group_id", gid).Error; err != nil {
		t.Fatal(err)
	}
	rows, err := dietrepo.FoodsAllWithVariantSiblings([]uint{a.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
	if len(rows[0].Variants) != 1 {
		t.Fatalf("variants %+v", rows[0].Variants)
	}
}

func TestCompositeFoodCreateAndByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	f := testutil.SeedFood(t, db, "Egg", testutil.DefaultMacros())
	cf := &models.CompositeFood{
		Name: "Scramble",
		Items: []models.CompositeFoodItem{{
			FoodID: f.ID,
			Amount: 2,
		}},
	}
	id, err := dietrepo.CompositeFoodCreate(cf)
	if err != nil {
		t.Fatal(err)
	}
	all, err := dietrepo.CompositeFoodsAll()
	if err != nil || len(all) != 1 {
		t.Fatalf("all %+v err %v", all, err)
	}
	loaded, err := dietrepo.CompositeFoodByID(id)
	if err != nil || loaded.Name != "Scramble" {
		t.Fatalf("loaded %+v err %v", loaded, err)
	}
}

func TestUpdatePlanMacros(t *testing.T) {
	db := testutil.SetupTestDB(t)
	p := testutil.SeedPlan(t, db, "Cut")
	updated, err := dietrepo.UpdatePlanMacros(p.ID, 1800, 140, 25, 150, 55)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Calories != 1800 || updated.Protein != 140 {
		t.Fatalf("got %+v", updated)
	}
}

func TestUpdatePlanMacros_notFound(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := dietrepo.UpdatePlanMacros(9999, 1, 1, 1, 1, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
