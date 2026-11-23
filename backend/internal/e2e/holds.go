//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// A hold catches the next request matching Method + Path and keeps it waiting until the test
// decides: release it, fail it through a DB fault, or drop the connection.
//
//	phase "before": the request waits before the handler runs (nothing written yet). Use it to
//	                assert pending/optimistic UI, or to fail the request for real.
//	phase "after":  the handler runs and commits; the response is held back. Use it for "saved on
//	                the server, but the client never heard back" (drop), the data-loss case that
//	                client_id idempotency exists for.
type holdSpec struct {
	Method string `json:"method" binding:"required"`
	// Path is a route template ("/workout/sets/:id") or a concrete path ("/workout/sets/12").
	Path      string `json:"path" binding:"required"`
	Phase     string `json:"phase"`     // before (default) | after
	TimeoutMs int    `json:"timeoutMs"` // auto-release after this long (default 15s); reported as timedOut
}

type decision struct {
	action string // release | fail | drop
	fault  *fault
}

type hold struct {
	holdSpec
	ID       string
	claimed  bool // a request matched it
	arrived  chan struct{}
	decided  chan decision
	once     sync.Once
	mu       sync.Mutex
	timedOut bool
	decidedB bool
}

const defaultHoldTimeout = 15 * time.Second

func (s *state) addHold(spec holdSpec) (string, error) {
	spec.Method = strings.ToUpper(spec.Method)
	switch spec.Phase {
	case "":
		spec.Phase = "before"
	case "before", "after":
	default:
		return "", fmt.Errorf("phase %q: want before or after", spec.Phase)
	}
	if !s.routeExists(spec.Method, spec.Path) {
		return "", fmt.Errorf("no route matches %s %s", spec.Method, spec.Path)
	}
	h := &hold{holdSpec: spec, arrived: make(chan struct{}), decided: make(chan decision, 1)}
	s.mu.Lock()
	defer s.mu.Unlock()
	h.ID = s.id()
	s.holds = append(s.holds, h)
	return h.ID, nil
}

// routeExists checks the pattern against the app's real routes, so a typo fails when armed rather
// than never matching.
func (s *state) routeExists(method, path string) bool {
	for _, r := range s.router.Routes() {
		if r.Method == method && (r.Path == path || templateMatches(r.Path, path)) {
			return true
		}
	}
	return false
}

// templateMatches reports whether a concrete path fits a gin route template (":param" segments).
func templateMatches(template, path string) bool {
	ts := strings.Split(strings.Trim(template, "/"), "/")
	ps := strings.Split(strings.Trim(path, "/"), "/")
	if len(ts) != len(ps) {
		return false
	}
	for i, t := range ts {
		if !strings.HasPrefix(t, ":") && t != ps[i] {
			return false
		}
	}
	return true
}

func (h *hold) matches(c *gin.Context) bool {
	if h.claimed || h.Method != c.Request.Method {
		return false
	}
	return h.Path == c.FullPath() || h.Path == c.Request.URL.Path
}

func (h *hold) resolve(d decision) {
	h.once.Do(func() {
		h.mu.Lock()
		h.decidedB = true
		h.mu.Unlock()
		h.decided <- d
	})
}

func (h *hold) view() holdView {
	h.mu.Lock()
	defer h.mu.Unlock()
	arrived := false
	select {
	case <-h.arrived:
		arrived = true
	default:
	}
	return holdView{h.ID, h.Method, h.Path, h.Phase, arrived, h.decidedB, h.timedOut}
}

func (s *state) claim(c *gin.Context) *hold {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, h := range s.holds {
		if h.matches(c) {
			h.claimed = true
			return h
		}
	}
	return nil
}

func (s *state) waitArrived(id string, wait time.Duration) (bool, error) {
	h, err := s.findHold(id)
	if err != nil {
		return false, err
	}
	select {
	case <-h.arrived:
		return true, nil
	case <-time.After(wait):
		return false, nil
	}
}

func (s *state) decide(id string, d decision) error {
	h, err := s.findHold(id)
	if err != nil {
		return err
	}
	switch d.action {
	case "release", "drop":
	case "fail":
		if h.Phase != "before" {
			return fmt.Errorf("fail needs phase before (the handler has already run)")
		}
		if _, err := s.addFault(*d.fault); err != nil {
			return err
		}
	default:
		return fmt.Errorf("action %q: want release, fail or drop", d.action)
	}
	h.resolve(d)
	return nil
}

// await blocks until the test decides, or the safety timeout releases the request.
func (h *hold) await() decision {
	timeout := defaultHoldTimeout
	if h.TimeoutMs > 0 {
		timeout = time.Duration(h.TimeoutMs) * time.Millisecond
	}
	select {
	case d := <-h.decided:
		return d
	case <-time.After(timeout):
		h.mu.Lock()
		h.timedOut = true
		h.mu.Unlock()
		log.Printf("e2e: hold %s (%s %s) timed out after %s, releasing", h.ID, h.Method, h.Path, timeout)
		return decision{action: "release"}
	}
}

func (s *state) holdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isE2EPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		h := s.claim(c)
		if h == nil {
			c.Next()
			return
		}
		if h.Phase == "before" {
			close(h.arrived)
			if d := h.await(); d.action == "drop" {
				dropConnection(c, c.Writer)
				return
			}
			c.Next() // "fail" armed its fault in decide(); the handler hits it for real
			return
		}

		// after: run the handler into a buffer, then hold the response.
		orig := c.Writer
		buf := &bufferedWriter{ResponseWriter: orig, status: http.StatusOK}
		c.Writer = buf
		c.Next()
		c.Writer = orig
		close(h.arrived)
		if d := h.await(); d.action == "drop" {
			dropConnection(c, orig)
			return
		}
		buf.flush()
	}
}

// dropConnection closes the TCP connection without a response: the client sees a network error.
func dropConnection(c *gin.Context, w gin.ResponseWriter) {
	c.Abort()
	conn, _, err := w.Hijack()
	if err != nil {
		log.Printf("e2e: drop: hijack failed: %v", err)
		return
	}
	_ = conn.Close()
}

// bufferedWriter collects a handler's response so it can be sent (or dropped) later.
type bufferedWriter struct {
	gin.ResponseWriter
	status  int
	written bool
	body    bytes.Buffer
}

func (w *bufferedWriter) WriteHeader(code int) {
	if code > 0 && !w.written {
		w.status = code
	}
}
func (w *bufferedWriter) WriteHeaderNow() { w.written = true }
func (w *bufferedWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.body.Write(b)
}
func (w *bufferedWriter) WriteString(s string) (int, error) {
	w.written = true
	return w.body.WriteString(s)
}
func (w *bufferedWriter) Status() int   { return w.status }
func (w *bufferedWriter) Size() int     { return w.body.Len() }
func (w *bufferedWriter) Written() bool { return w.written }
func (w *bufferedWriter) Flush()        {}

func (w *bufferedWriter) flush() {
	w.ResponseWriter.WriteHeader(w.status)
	w.ResponseWriter.WriteHeaderNow()
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
}
