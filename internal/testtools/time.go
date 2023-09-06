package tt

import (
	"testing"
	"time"
)

func ParseTime(t *testing.T, timestr string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, timestr)
	AssertNoErr(t, err)
	return parsedTime
}

func MockTimeAfter(
	timeController func(triggerCh chan time.Time, waitCh chan time.Duration),
) func(d time.Duration) <-chan time.Time {
	waitCh := make(chan time.Duration, 10)
	triggerCh := make(chan time.Time)

	go timeController(triggerCh, waitCh)

	return func(d time.Duration) <-chan time.Time {
		waitCh <- d
		return triggerCh
	}
}
