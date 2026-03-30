package debugging

import (
	"time"
)

// PossibleBlocking runs the function f in a goroutine and returns the result.
// If the function f is blocking, the test will fail.
func PossibleBlocking[T any](f func() T, timeout time.Duration, trigger func()) T {
	d := make(chan T)

	go func() {
		d <- f()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			trigger()
		case v := <-d:
			return v
		}
	}
}
