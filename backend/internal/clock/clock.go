// Package clock is the server's source of "now". Domain code calls clock.Now() instead of
// time.Now() so the e2e suite can move the server's clock (see internal/e2e). In every other build
// nothing calls Set and Now is time.Now.
//
// Use time.Now directly only for durations and uniqueness (timers, benchmarks, nonces), where a
// moved clock would be wrong.
package clock

import (
	"sync/atomic"
	"time"
)

// offset is added to the real time. It is an offset rather than a frozen instant so the server's
// clock keeps ticking, the same way Playwright's page.clock.install() does in the browser.
var offset atomic.Int64

// Now returns the current time, shifted by the offset set with Set.
func Now() time.Time {
	if d := offset.Load(); d != 0 {
		return time.Now().Add(time.Duration(d))
	}
	return time.Now()
}

// Set makes Now return t at this instant and tick on from there. Test-only (internal/e2e).
func Set(t time.Time) { offset.Store(int64(time.Until(t))) }

// Reset makes Now the real time again.
func Reset() { offset.Store(0) }
