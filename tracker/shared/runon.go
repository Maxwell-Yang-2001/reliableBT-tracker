package shared

import (
	"time"
)

// RunOn will run the given function at the exact tick of the duration.
// For example `RunOn(1*time.Second, func)` would run on the second every second
func RunOn(duration time.Duration, run func()) {
	nextTick := time.Now().Truncate(duration)
	for {
		nextTick = nextTick.Add(duration)
		time.Sleep(time.Until(nextTick))
		run()
	}
}
