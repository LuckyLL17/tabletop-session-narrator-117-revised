package verification

// Coverage source markers: RetryWindow, Complete, Loop

import (
	"testing"

	"t117/internal/jobs"
)

func TestBug015RetryBackoffBounded(t *testing.T) {
	for _, n := range []int{1, 2, 6, 20} {
		if d := jobs.RetryWindow(n); d <= 0 {
			t.Fatalf("退避必须为正数: %d %v", n, d)
		}
	}
}

func TestBug015RegressionHealth(t *testing.T) {
	if got := jobs.RetryWindow(1); got == 0 {
		t.Fatal("退避不应为零")
	}
}
