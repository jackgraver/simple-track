package database

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ownedThing struct {
	gorm.Model
	Owner
	Name string
}

type catalogThing struct {
	gorm.Model
	Creator
	Name string
}

func ownerKitDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if d, err := db.DB(); err == nil {
		d.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(&ownedThing{}, &catalogThing{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func as(db *gorm.DB, uid uint) context.Context {
	return WithIdentity(WithDB(context.Background(), db), Identity{UserID: uid})
}

func TestOwnerStampsCreates(t *testing.T) {
	db := ownerKitDB(t)
	ctx := as(db, 7)
	row := ownedThing{Name: "a"}
	if err := From(ctx).Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	batch := []ownedThing{{Name: "b"}, {Name: "c"}}
	if err := From(ctx).Create(&batch).Error; err != nil {
		t.Fatal(err)
	}
	cat := catalogThing{Name: "food"}
	if err := From(ctx).Create(&cat).Error; err != nil {
		t.Fatal(err)
	}
	if row.UserID != 7 || batch[0].UserID != 7 || batch[1].UserID != 7 || cat.CreatedBy != 7 {
		t.Fatalf("not stamped: %d %d %d %d", row.UserID, batch[0].UserID, batch[1].UserID, cat.CreatedBy)
	}
}

func TestOwnerRefusesCreateForSomeoneElse(t *testing.T) {
	db := ownerKitDB(t)
	row := ownedThing{Owner: Owner{UserID: 8}, Name: "forged"}
	err := From(as(db, 7)).Create(&row).Error
	if err == nil || !strings.Contains(err.Error(), "user_id 8 set by hand") {
		t.Fatalf("err = %v, want a refusal", err)
	}
	cat := catalogThing{Creator: Creator{CreatedBy: 8}}
	if err := From(as(db, 7)).Create(&cat).Error; err == nil {
		t.Fatal("created_by for someone else accepted")
	}
}

func TestOwnerWithoutIdentity(t *testing.T) {
	db := ownerKitDB(t)
	ctx := WithDB(context.Background(), db)
	if err := From(ctx).Create(&ownedThing{Name: "x"}).Error; !errors.Is(err, ErrNoIdentity) {
		t.Fatalf("err = %v, want ErrNoIdentity", err)
	}
	// Seeds and tools (owner role, no request) set the owner themselves.
	if err := From(ctx).Create(&ownedThing{Owner: Owner{UserID: 3}, Name: "seeded"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := From(ctx).Create(&catalogThing{Name: "x"}).Error; !errors.Is(err, ErrNoIdentity) {
		t.Fatalf("catalog err = %v, want ErrNoIdentity", err)
	}
}

func TestOwnedScope(t *testing.T) {
	db := ownerKitDB(t)
	for uid, names := range map[uint][]string{1: {"a1", "a2"}, 2: {"b1"}} {
		for _, n := range names {
			if err := From(as(db, uid)).Create(&ownedThing{Name: n}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	var mine []ownedThing
	ctx := as(db, 2)
	if err := From(ctx).Scopes(Owned(ctx)).Find(&mine).Error; err != nil {
		t.Fatal(err)
	}
	if len(mine) != 1 || mine[0].Name != "b1" {
		t.Fatalf("user 2 sees %+v", mine)
	}

	// Someone else's row reads, updates and deletes as not found.
	var a1 ownedThing
	if err := db.Where("name = ?", "a1").First(&a1).Error; err != nil {
		t.Fatal(err)
	}
	if err := From(ctx).Scopes(Owned(ctx)).First(&ownedThing{}, a1.ID).Error; !IsNotFound(err) {
		t.Fatalf("lookup of user 1's row as 2: %v", err)
	}
	if err := Affected(From(ctx).Scopes(Owned(ctx)).Delete(&ownedThing{}, a1.ID)); !IsNotFound(err) {
		t.Fatalf("delete of user 1's row as 2: %v", err)
	}
	if err := Affected(From(ctx).Model(&ownedThing{}).Scopes(Owned(ctx)).Where("id = ?", a1.ID).Update("name", "x")); !IsNotFound(err) {
		t.Fatalf("update of user 1's row as 2: %v", err)
	}

	// Soft-deleted rows stay scoped under Unscoped.
	if err := Affected(From(as(db, 1)).Scopes(Owned(as(db, 1))).Delete(&ownedThing{}, a1.ID)); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := From(ctx).Unscoped().Model(&ownedThing{}).Scopes(Owned(ctx)).Count(&n).Error; err != nil || n != 1 {
		t.Fatalf("unscoped count as 2: %d, %v", n, err)
	}

	// Joins: the user_id column is qualified with the statement's table.
	var joined []ownedThing
	err := From(ctx).Scopes(Owned(ctx)).
		Joins("JOIN catalog_things ON catalog_things.name = owned_things.name").Find(&joined).Error
	if err != nil {
		t.Fatalf("scoped join: %v", err)
	}
}

func TestOwnedWithoutIdentityFails(t *testing.T) {
	db := ownerKitDB(t)
	ctx := WithDB(context.Background(), db)
	var rows []ownedThing
	if err := From(ctx).Scopes(Owned(ctx)).Find(&rows).Error; !errors.Is(err, ErrNoIdentity) {
		t.Fatalf("err = %v, want ErrNoIdentity", err)
	}
}

func TestCurrentIdentityFromRequest(t *testing.T) {
	db := ownerKitDB(t)
	ctx, req := WithTx(context.Background(), db)
	defer req.Rollback()
	if _, ok := CurrentIdentity(ctx); ok {
		t.Fatal("identity before SetIdentity")
	}
	if err := SetIdentity(ctx, Identity{UserID: 5, IsAdmin: true}); err != nil {
		t.Fatal(err)
	}
	// The request's identity wins over a WithIdentity value.
	id, ok := CurrentIdentity(WithIdentity(ctx, Identity{UserID: 9}))
	if !ok || id.UserID != 5 || !id.IsAdmin {
		t.Fatalf("identity = %+v, %v", id, ok)
	}
	row := ownedThing{Name: "r"}
	if err := From(ctx).Create(&row).Error; err != nil || row.UserID != 5 {
		t.Fatalf("create in request: %v, user_id %d", err, row.UserID)
	}
}

func TestLockKey(t *testing.T) {
	db := ownerKitDB(t)
	key, err := LockKey(as(db, 42), "workout_logs.date", "2026-01-15")
	if err != nil || key != "u42:workout_logs.date:2026-01-15" {
		t.Fatalf("key %q, %v", key, err)
	}
	if _, err := LockKey(WithDB(context.Background(), db), "x"); !errors.Is(err, ErrNoIdentity) {
		t.Fatalf("no identity: %v", err)
	}
	if _, err := LockKey(as(db, 1)); err == nil {
		t.Fatal("empty key accepted")
	}
	if err := XactLock(as(db, 1), From(as(db, 1)), "x"); err != nil {
		t.Fatalf("XactLock on SQLite: %v", err)
	}
}
