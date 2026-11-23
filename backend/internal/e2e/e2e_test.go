//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type widget struct {
	ID   uint
	Name string
}

// harness: an app with one table and two routes, plus the e2e state (no /__e2e routes, no reset:
// those need Postgres and are covered by the Playwright suite).
func harness(t *testing.T) (*state, *gorm.DB, *httptest.Server) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&widget{}); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	s := newState(r, db)
	if err := s.registerFaultHooks(db); err != nil {
		t.Fatal(err)
	}
	r.Use(s.holdMiddleware())
	r.POST("/widgets", func(c *gin.Context) {
		w := widget{Name: "w"}
		if err := db.Create(&w).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, w)
	})
	r.GET("/widgets/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"id": c.Param("id")}) })
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return s, db, srv
}

func post(t *testing.T, url string) (int, string, error) {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewReader(nil))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b), nil
}

func count(t *testing.T, db *gorm.DB) int64 {
	var n int64
	db.Model(&widget{}).Count(&n)
	return n
}

func TestFaultFailsRealHandlerOnce(t *testing.T) {
	s, db, srv := harness(t)
	if _, err := s.addFault(fault{Table: "widgets", Op: "create", Times: 1}); err != nil {
		t.Fatal(err)
	}
	code, body, _ := post(t, srv.URL+"/widgets")
	if code != 500 || !bytes.Contains([]byte(body), []byte("e2e fault: create widgets")) {
		t.Fatalf("first call: %d %s", code, body)
	}
	if code, _, _ := post(t, srv.URL+"/widgets"); code != 201 {
		t.Fatalf("second call: %d, want 201 (times: 1)", code)
	}
	if n := count(t, db); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
	if got := s.snapshot()["faults"].([]faultView)[0].Hits; got != 1 {
		t.Fatalf("hits = %d", got)
	}
}

func TestFaultPgCode(t *testing.T) {
	s, db, _ := harness(t)
	s.addFault(fault{Table: "widgets", Op: "create", Code: "23505"})
	err := db.Create(&widget{Name: "x"}).Error
	var pg *pgconn.PgError
	if !errors.As(err, &pg) || pg.Code != "23505" {
		t.Fatalf("err = %v, want PgError 23505", err)
	}
}

func TestArmValidation(t *testing.T) {
	s, _, _ := harness(t)
	if _, err := s.addFault(fault{Table: "nope", Op: "create"}); err == nil {
		t.Error("unknown table accepted")
	}
	if _, err := s.addFault(fault{Table: "widgets", Op: "insert"}); err == nil {
		t.Error("unknown op accepted")
	}
	if _, err := s.addHold(holdSpec{Method: "POST", Path: "/widget"}); err == nil {
		t.Error("unknown route accepted")
	}
	for _, p := range []string{"/widgets/:id", "/widgets/7"} {
		if _, err := s.addHold(holdSpec{Method: "get", Path: p}); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
}

type result struct {
	code int
	err  error
}

func async(t *testing.T, url string) chan result {
	ch := make(chan result, 1)
	go func() {
		code, _, err := post(t, url)
		ch <- result{code, err}
	}()
	return ch
}

func arrive(t *testing.T, s *state, id string) {
	t.Helper()
	if ok, _ := s.waitArrived(id, 2*time.Second); !ok {
		t.Fatal("request never arrived")
	}
}

func TestHoldBeforeRelease(t *testing.T) {
	s, db, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets"})
	ch := async(t, srv.URL+"/widgets")
	arrive(t, s, id)
	if n := count(t, db); n != 0 {
		t.Fatalf("handler ran before release (rows = %d)", n)
	}
	s.decide(id, decision{action: "release"})
	if r := <-ch; r.code != 201 {
		t.Fatalf("got %+v", r)
	}
}

func TestHoldBeforeFail(t *testing.T) {
	s, db, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets"})
	ch := async(t, srv.URL+"/widgets")
	arrive(t, s, id)
	if err := s.decide(id, decision{action: "fail", fault: &fault{Table: "widgets", Op: "create", Times: 1}}); err != nil {
		t.Fatal(err)
	}
	if r := <-ch; r.code != 500 {
		t.Fatalf("got %+v, want 500", r)
	}
	if n := count(t, db); n != 0 {
		t.Fatalf("rows = %d", n)
	}
}

func TestHoldAfterDropCommitsButClientErrors(t *testing.T) {
	s, db, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets", Phase: "after"})
	ch := async(t, srv.URL+"/widgets")
	arrive(t, s, id)
	if n := count(t, db); n != 1 {
		t.Fatalf("rows = %d, want the write committed before the response", n)
	}
	s.decide(id, decision{action: "drop"})
	if r := <-ch; r.err == nil {
		t.Fatalf("got %+v, want a network error", r)
	}
	if err := s.decide(id, decision{action: "fail", fault: &fault{Table: "widgets", Op: "create"}}); err == nil {
		t.Error("fail accepted for an after-phase hold")
	}
}

func TestHoldAfterRelease(t *testing.T) {
	s, _, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets", Phase: "after"})
	ch := async(t, srv.URL+"/widgets")
	arrive(t, s, id)
	s.decide(id, decision{action: "release"})
	r := <-ch
	if r.code != 201 {
		t.Fatalf("got %+v", r)
	}
}

func TestHoldTimeoutReleases(t *testing.T) {
	s, _, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets", TimeoutMs: 50})
	if r := <-async(t, srv.URL+"/widgets"); r.code != 201 {
		t.Fatalf("got %+v", r)
	}
	h, _ := s.findHold(id)
	if !h.view().TimedOut {
		t.Error("timedOut not recorded")
	}
}

func TestClearReleasesPendingHolds(t *testing.T) {
	s, _, srv := harness(t)
	id, _ := s.addHold(holdSpec{Method: "POST", Path: "/widgets"})
	ch := async(t, srv.URL+"/widgets")
	arrive(t, s, id)
	s.clear()
	if r := <-ch; r.code != 201 {
		t.Fatalf("got %+v", r)
	}
	b, _ := json.Marshal(s.snapshot())
	if string(b) != `{"faults":[],"holds":[]}` {
		t.Fatalf("state after clear: %s", b)
	}
}
