package clock

import (
	"testing"
	"time"
)

func TestSetAndReset(t *testing.T) {
	t.Cleanup(Reset)
	target := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	Set(target)
	if got := Now(); got.Sub(target).Abs() > time.Second {
		t.Fatalf("Now() = %v, want ~%v", got, target)
	}
	Reset()
	if got := Now(); time.Since(got).Abs() > time.Second {
		t.Fatalf("after Reset Now() = %v, want real time", got)
	}
}
