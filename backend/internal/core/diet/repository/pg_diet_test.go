package dietrepo_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/url"
	"sync"
	"testing"

	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/testutil"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/pgtest"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// pgDietDB creates a fresh, fully migrated database on TEST_DATABASE_URL (skipped when unset) and
// returns it. Same convention as internal/migrate's tests.
func pgDietDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Skipped without Postgres, or fails when TEST_DATABASE_REQUIRED is set (CI).
	adminDB, adminURL := pgtest.Admin(t)
	admin := adminURL.String()
	name := fmt.Sprintf("diet_test_%d", rand.Uint32())
	if _, err := adminDB.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	sqlDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	gdb, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	// The user every seeded row belongs to (testutil.UserID is the first identity value).
	if err := gdb.Exec(`INSERT INTO users (username, password, email, is_admin, created_at)
		VALUES ('owner', 'x', 'o@x', true, now())`).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if d, err := gdb.DB(); err == nil {
			d.Close()
		}
		sqlDB.Close()
		adminDB.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	return gdb
}

// inRequest runs fn the way a request does: in its own lazy request transaction, committed when
// fn succeeds.
func inRequest(db *gorm.DB, fn func(ctx context.Context) error) error {
	ctx, req := database.WithTx(context.Background(), db)
	if err := database.SetIdentity(ctx, database.Identity{UserID: testutil.UserID}); err != nil {
		return err
	}
	if err := fn(ctx); err != nil {
		req.Rollback()
		return err
	}
	return req.Commit()
}

// parallel runs fn n times at once and fails on any error.
func parallel(t *testing.T, n int, fn func(i int) error) {
	t.Helper()
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs <- fn(i)
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Review findings 2 and 3: concurrent replays of diet creates and of log-planned must not hit
// the unique indexes (which surfaced as 500s).
func TestPostgres_dietConcurrentReplays(t *testing.T) {
	db := pgDietDB(t)
	plan := testutil.SeedPlan(t, db, "P")
	day := testutil.SeedDay(t, db, testutil.Today(), plan.ID)
	food := testutil.SeedFood(t, db, "Oats", testutil.DefaultMacros())
	saved := testutil.SeedSavedMeal(t, db, "Breakfast", food.ID)

	// Quick log: one request retried 12 times concurrently.
	qlID := "22222222-2222-4222-8222-222222222222"
	parallel(t, 12, func(i int) error {
		return inRequest(db, func(ctx context.Context) error {
			_, err := dietrepo.QuickLogMeal(ctx, dietrepo.QuickLogParams{
				DisplayName: "Bar", FoodRowName: fmt.Sprintf("Bar [ql-%d]", i), Calories: 200, ClientID: qlID,
			})
			return err
		})
	})
	var n int64
	db.Model(&models.Meal{}).Where("client_id = ?", qlID).Count(&n)
	if n != 1 {
		t.Fatalf("quick log: want 1 meal, got %d", n)
	}

	// Planned from saved: same, then a replay after the planned meal was (soft) deleted.
	pmID := "33333333-3333-4333-8333-333333333333"
	addFromSaved := func(int) error {
		meal := models.Meal{Name: saved.Name, Items: []models.MealItem{{FoodID: food.ID, Amount: 1}}}
		return inRequest(db, func(ctx context.Context) error {
			return dietrepo.AddPlannedMealFromSavedMeal(ctx, 0, saved.ID, &meal, pmID)
		})
	}
	parallel(t, 12, addFromSaved)
	var pm models.PlannedMeal
	if err := db.Where("client_id = ?", pmID).First(&pm).Error; err != nil {
		t.Fatal(err)
	}
	db.Model(&models.PlannedMeal{}).Where("client_id = ?", pmID).Count(&n)
	if n != 1 {
		t.Fatalf("from saved: want 1 planned meal, got %d", n)
	}
	if err := db.Delete(&pm).Error; err != nil {
		t.Fatal(err)
	}
	if err := addFromSaved(0); err != nil {
		t.Fatalf("replay after delete: %v", err)
	}

	// Log planned after an unlog (logged stays true): concurrent logs leave one live day log.
	meal := testutil.SeedMeal(t, db, "Lunch", food.ID, 1)
	testutil.SeedPlannedMeal(t, db, day.ID, meal.ID, 0, true)
	for round := 0; round < 5; round++ {
		parallel(t, 8, func(int) error {
			return inRequest(db, func(ctx context.Context) error { return dietrepo.LogPlannedMeal(ctx, day.ID, meal.ID) })
		})
		db.Model(&models.DayLog{}).Where("day_id = ? AND meal_id = ?", day.ID, meal.ID).Count(&n)
		if n != 1 {
			t.Fatalf("round %d: want 1 live day log, got %d", round, n)
		}
		if err := inRequest(db, func(ctx context.Context) error { return dietrepo.DeleteLoggedMeal(ctx, day.ID, meal.ID) }); err != nil {
			t.Fatal(err)
		}
	}
}
