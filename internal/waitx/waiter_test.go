package waitx

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeChecker struct {
	name      string
	failUntil int
	attempts  int
}

func (f *fakeChecker) Name() string {
	return f.name
}

func (f *fakeChecker) Check(_ context.Context) error {
	f.attempts++
	if f.attempts <= f.failUntil {
		return errors.New("not ready")
	}
	return nil
}

func TestWaitContextSuccess(t *testing.T) {
	chk := &fakeChecker{name: "fake", failUntil: 2}
	opts := DefaultWaitOptions()
	opts.Timeout = 2 * time.Second
	opts.Interval = 5 * time.Millisecond
	opts.RetryStrategy = LinearRetry{}

	err := WaitContext(context.Background(), chk, opts)
	if err != nil {
		t.Fatalf("expected wait success, got error: %v", err)
	}
	if chk.attempts < 3 {
		t.Fatalf("expected at least 3 attempts, got %d", chk.attempts)
	}
}

func TestWaitContextTimeout(t *testing.T) {
	chk := &fakeChecker{name: "never-ready", failUntil: 100}
	opts := DefaultWaitOptions()
	opts.Timeout = 20 * time.Millisecond
	opts.Interval = 5 * time.Millisecond

	err := WaitContext(context.Background(), chk, opts)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitContextInvertCheck(t *testing.T) {
	chk := &fakeChecker{name: "invert", failUntil: 100}
	opts := DefaultWaitOptions()
	opts.Timeout = 100 * time.Millisecond
	opts.Interval = 5 * time.Millisecond
	opts.InvertCheck = true

	err := WaitContext(context.Background(), chk, opts)
	if err != nil {
		t.Fatalf("expected invert check success, got %v", err)
	}
}
