package database

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Middleware gives each request a lazy transaction on root (see From). Install it after any
// middleware that holds requests or buffers responses (the e2e holds), so a held request isn't
// sitting in an open transaction, and before the routes that use From.
//
// The transaction ends when the handler first writes its response, so a failed commit can still
// turn into a 500 instead of a success the client would trust: status < 400 commits, anything
// else rolls back. A handler that writes nothing ends it when it returns, and a panic rolls it back.
func Middleware(root *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, req := WithTx(c.Request.Context(), root)
		c.Request = c.Request.WithContext(ctx)
		orig := c.Writer
		w := &txWriter{ResponseWriter: orig, req: req, method: c.Request.Method, path: c.FullPath()}
		c.Writer = w
		defer func() {
			// Runs on a panic too (it keeps unwinding to gin's Recovery, which then writes to orig).
			c.Writer = orig
			if err := req.Rollback(); err != nil {
				log.Printf("database: rollback %s %s: %v", w.method, w.path, err)
			}
		}()
		c.Next()
		w.settle()
		if w.failed && !orig.Written() {
			w.writeFailure()
		}
	}
}

// txWriter ends the request transaction just before the first byte of the response goes out.
type txWriter struct {
	gin.ResponseWriter
	req     *Request
	method  string
	path    string
	settled bool
	failed  bool // commit failed: the real response is replaced by a 500
}

// settle commits (status < 400) or rolls back, once.
func (w *txWriter) settle() {
	if w.settled {
		return
	}
	w.settled = true
	if w.ResponseWriter.Status() >= http.StatusBadRequest {
		if err := w.req.Rollback(); err != nil {
			log.Printf("database: rollback %s %s: %v", w.method, w.path, err)
		}
		return
	}
	if err := w.req.Commit(); err != nil {
		log.Printf("database: commit %s %s: %v", w.method, w.path, err)
		w.failed = true
	}
}

func (w *txWriter) writeFailure() {
	h := w.ResponseWriter.Header()
	// Nothing the handler meant for its success may go out with the 500 (a session cookie for a
	// user row that was rolled back, a redirect, a cache validator).
	for _, k := range []string{"Set-Cookie", "Location", "ETag", "Last-Modified", "Content-Disposition", "Content-Length"} {
		h.Del(k)
	}
	h.Set("Content-Type", "application/json; charset=utf-8")
	w.ResponseWriter.WriteHeader(http.StatusInternalServerError)
	_, _ = w.ResponseWriter.WriteString(`{"error":"Could not save changes"}`)
}

func (w *txWriter) Write(b []byte) (int, error) {
	w.settle()
	if w.failed {
		if !w.ResponseWriter.Written() {
			w.writeFailure()
		}
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func (w *txWriter) WriteString(s string) (int, error) {
	w.settle()
	if w.failed {
		if !w.ResponseWriter.Written() {
			w.writeFailure()
		}
		return len(s), nil
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *txWriter) WriteHeaderNow() {
	w.settle()
	if w.failed {
		if !w.ResponseWriter.Written() {
			w.writeFailure()
		}
		return
	}
	w.ResponseWriter.WriteHeaderNow()
}

func (w *txWriter) Flush() {
	w.WriteHeaderNow()
	w.ResponseWriter.Flush()
}

func (w *txWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.settle()
	if w.failed {
		return nil, nil, errCommitFailed
	}
	return w.ResponseWriter.Hijack()
}

var errCommitFailed = errors.New("database: request transaction failed to commit")
