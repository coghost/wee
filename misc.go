package wee

import (
	"math"
	"math/rand"
	"time"
)

// FirstOrDefault
//
// return the first args value, if args not empty
// else return default value
func FirstOrDefault[T any](dft T, args ...T) T { //nolint: ireturn
	val := dft
	if len(args) > 0 {
		val = args[0]
	}

	return val
}

// RandFloatX1k returns a value between (min, max) * 1000
func RandFloatX1k(min, max float64) int {
	secToMill := 1000.0

	minI, maxI := int(math.Round(min*secToMill)), int(math.Round(max*secToMill))
	x1k := minI + rand.Intn(maxI-minI)

	return x1k
}

// RandSleep random sleep some seconds
//
// @returns milliseconds
func RandSleep(min, max float64, msg ...string) int {
	slept := RandFloatX1k(min, max)
	time.Sleep(time.Duration(slept) * time.Millisecond)

	return slept
}

// NewStringSlice
//
// raw: the raw string to be convert to slice
// fixStep: how many chars in each slice
// args: just add any one to enable random step mode
//
// return:
//
//	str slice with each has (~)maxLen chars
func NewStringSlice(raw string, fixStep int, randomStep ...bool) []string {
	rand.Seed(time.Now().Unix())

	var ret []string

	step := fixStep
	perStr := ""

	isRandStep := FirstOrDefault(false, randomStep...)

	for _, ch := range raw {
		perStr += string(ch)
		if len(perStr) >= step {
			ret = append(ret, perStr)
			perStr = ""

			if !isRandStep {
				continue
			}

			chance := rand.Float64()

			switch {
			case chance > 0.8:
				step = fixStep + 2
			case chance > 0.6:
				step = fixStep + 1
			case chance < 0.2:
				step = fixStep - 2
			case chance > 0.4:
				step = fixStep - 1
			default:
				step = fixStep
			}
		}
	}

	if perStr != "" {
		ret = append(ret, perStr)
	}

	return ret
}

func refineIndex(count, index int) int {
	if count == 0 {
		return -1
	}

	if index < 0 {
		index = count + index
	}

	index = min(index, count-1)

	return index
}
