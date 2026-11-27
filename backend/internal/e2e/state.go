//go:build e2e

package e2e

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// state holds the armed faults and holds. Each e2e worker has its own backend process and runs one
// test at a time, so plain process memory is enough; /__e2e/reset clears it.
type state struct {
	mu     sync.Mutex
	router *gin.Engine
	db     *gorm.DB // the server's pool (app role): fault hooks live here
	seedDB *gorm.DB // owner-role pool for reset/seed (TRUNCATE ... RESTART IDENTITY needs ownership)
	nextID int
	faults []*fault
	holds  []*hold
}

func newState(router *gin.Engine, db *gorm.DB) *state {
	return &state{router: router, db: db, seedDB: db}
}

func (s *state) id() string {
	s.nextID++
	return strconv.Itoa(s.nextID)
}

// clear drops every fault and releases any request still held, so nothing leaks into the next test.
func (s *state) clear() {
	s.mu.Lock()
	holds := s.holds
	s.faults, s.holds = nil, nil
	s.mu.Unlock()
	for _, h := range holds {
		h.resolve(decision{action: "release"})
	}
}

type faultView struct {
	ID    string `json:"id"`
	Table string `json:"table"`
	Op    string `json:"op"`
	Times int    `json:"times"`
	Hits  int    `json:"hits"`
}

type holdView struct {
	ID       string `json:"id"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Phase    string `json:"phase"`
	Arrived  bool   `json:"arrived"`
	Decided  bool   `json:"decided"`
	TimedOut bool   `json:"timedOut"`
}

func (s *state) snapshot() gin.H {
	s.mu.Lock()
	defer s.mu.Unlock()
	fs := make([]faultView, 0, len(s.faults))
	for _, f := range s.faults {
		fs = append(fs, faultView{f.ID, f.Table, f.Op, f.Times, f.Hits})
	}
	hs := make([]holdView, 0, len(s.holds))
	for _, h := range s.holds {
		hs = append(hs, h.view())
	}
	return gin.H{"faults": fs, "holds": hs}
}

func (s *state) findHold(id string) (*hold, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, h := range s.holds {
		if h.ID == id {
			return h, nil
		}
	}
	return nil, fmt.Errorf("no hold %q", id)
}
