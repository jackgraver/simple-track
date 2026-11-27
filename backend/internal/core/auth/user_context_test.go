package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserID_noUser(t *testing.T) {
	if id, ok := UserID(context.Background()); ok || id != 0 {
		t.Fatalf("got %d %v, want 0 false", id, ok)
	}
	// A zero ID never counts as a user.
	if _, ok := UserID(WithUser(context.Background(), User{Username: "ghost"})); ok {
		t.Fatal("user with ID 0 was accepted")
	}
}

func TestUserID_fromContextAndGinContext(t *testing.T) {
	ctx := WithUser(context.Background(), User{ID: 7, Username: "alice", IsAdmin: true})
	if id, ok := UserID(ctx); !ok || id != 7 {
		t.Fatalf("got %d %v", id, ok)
	}
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	gc.Request = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	if u := MustUser(gc); u.ID != 7 || !u.IsAdmin {
		t.Fatalf("via gin.Context: got %+v", u)
	}
}

func TestMustUser_panicsWithoutUser(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustUser did not panic")
		}
	}()
	MustUser(context.Background())
}
