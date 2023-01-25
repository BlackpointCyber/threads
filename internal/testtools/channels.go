package tt

import (
	"testing"
	"time"
)

func AssertDone(t *testing.T, timeout time.Duration, doneCh chan struct{}) {
	select {
	case <-time.After(timeout):
		t.Fatalf("the done channel was not closed even after %v", timeout)
	case <-doneCh:
	}
}

func AssertNotDone(t *testing.T, doneCh chan struct{}) {
	select {
	case <-doneCh:
		t.Fatalf("the done channel should be open but was closed")
	default:
	}
}
