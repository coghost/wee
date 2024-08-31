//nolint:mnd
package wee

import (
	"math"
	"math/rand"
	"time"
)

// RandFloatX1k returns a random integer value between (min * 1000) and (max * 1000).
//
// Parameters:
//   - min: The minimum value of the range (in seconds).
//   - max: The maximum value of the range (in seconds).
//
// Returns:
//   - An integer representing milliseconds, randomly chosen between min*1000 and max*1000.
//
// If min equals max, it returns min * 1000 as an integer.
// The function uses math.Round to ensure accurate conversion from seconds to milliseconds.
func RandFloatX1k(min, max float64) int {
	secToMill := 1000.0
	if min == max {
		return int(secToMill * min)
	}

	minI, maxI := int(math.Round(min*secToMill)), int(math.Round(max*secToMill))
	x1k := minI + rand.Intn(maxI-minI)

	return x1k
}

// RandSleep sleeps for a random duration between min and max seconds.
// It returns the actual sleep duration in milliseconds.
func RandSleep(min, max float64, msg ...string) int {
	slept := RandFloatX1k(min, max)
	time.Sleep(time.Duration(slept) * time.Millisecond)

	return slept
}

// SleepN sleeps for a random duration between n and n*1.1 seconds.
// It returns the actual sleep duration in milliseconds.
func SleepN(n float64) int {
	n1 := n * 1.1
	return RandSleep(n, n1)
}

// RandSleepNap sleeps for a random duration between 1 and 2 seconds.
// It returns the actual sleep duration in milliseconds.
func RandSleepNap() (sleepMills int) {
	return RandSleep(1.0, 2.0)
}

// RandSleepPT1Ms sleeps for a random duration between 0.01 and 0.1 milliseconds.
// It returns the actual sleep duration in milliseconds.
func RandSleepPT1Ms() (sleepMills int) {
	return RandSleep(0.01, 0.1)
}

// SleepPT100Ms sleeps 0.1~0.2s
func SleepPT100Ms() {
	RandSleep(0.1, 0.2)
}

// SleepPT250Ms sleeps 0.25~0.3s
func SleepPT250Ms() {
	RandSleep(0.25, 0.3)
}

// SleepPT500Ms sleeps 0.5~0.6s
func SleepPT500Ms() {
	RandSleep(0.5, 0.6)
}
