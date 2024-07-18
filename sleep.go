//nolint:mnd
package wee

import "time"

// RandSleep random sleep some seconds
//
// @returns milliseconds
func RandSleep(min, max float64, msg ...string) int {
	slept := RandFloatX1k(min, max)
	time.Sleep(time.Duration(slept) * time.Millisecond)

	return slept
}

// SleepN sleeps n~n*1.1s
func SleepN(n float64) int {
	n1 := n * 1.1
	return RandSleep(n, n1)
}

// RandSleepPT1Ms sleeps rand(1s~2s)
func RandSleepNap() (sleepMills int) {
	return RandSleep(1.0, 2.0)
}

// RandSleepPT1Ms sleeps rand(0.01ms~0.1ms)
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
