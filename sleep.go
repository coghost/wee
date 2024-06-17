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

// RandSleepPT1Ms sleeps rand(1s~2s)
func RandSleepNap() (sleepMills int) {
	return RandSleep(1.0, 2.0) //nolint:mnd
}

// RandSleepPT1Ms sleeps rand(0.01ms~0.1ms)
func RandSleepPT1Ms() (sleepMills int) {
	return RandSleep(0.01, 0.1) //nolint:mnd
}
