package testutils

import (
	"testing"
	"time"
)

func TestTester_After(t *testing.T) {
	t0 := time.Now()
	t1 := t0.Add(10 * time.Millisecond)
	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckAfter(t1, t0)
	})

	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckAfter(t0, t0, 5*time.Millisecond)
	})

	ensureNotFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckAfter(t0, t1)
	})
}

func TestTester_Before(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-10 * time.Millisecond) // 10ms earlier
	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckBefore(earlier, now)
	})

	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckBefore(now, now, -5*time.Millisecond) // only 5ms earlier
	})

	ensureNotFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckBefore(now, earlier)
	})
}
func TestTester_Fatalf(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.Fatalf("The pope's hat isn't funny")
	})
}

func TestTester_CheckTruef(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckTruef(false, "Oh no %s %s", "Blistering", "Barnacles")
	})
	ensureNotFailed(t, func(ft *testing.T) {
		tt := NewTester(ft)
		tt.CheckTruef(true, "Oh no %s %s", "Blistering", "Barnacles")
	})
}
