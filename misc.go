package wee

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/spf13/cast"
)

// FirstOrDefault
//
// return the first args value, if args not empty
// else return default value
func FirstOrDefault[T any](dft T, args ...T) T { //nolint:ireturn
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
	rand.NewSource(time.Now().Unix())

	var ret []string

	step := fixStep
	perStr := ""

	isRandStep := FirstOrDefault(false, randomStep...)

	//nolint:mnd
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

// Stringify returns a string representation
func Stringify(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func MustStringify(data interface{}) string {
	b, err := Stringify(data)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func MustStrToFloat(raw string, keptChars string) float64 {
	v := StrToNumChars(raw, keptChars)
	return cast.ToFloat64(v)
}

func StrToNumChars(raw string, keptChars string) string {
	chars := "[0-9" + keptChars + "]+"
	re := regexp.MustCompile(chars)
	c := re.FindAllString(raw, -1)
	r := strings.Join(c, "")

	return r
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsZeroVal check if any type is its zero value
func IsZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// StrAorB returns the first non empty string.
func StrAorB(a, b string) string {
	if a != "" {
		return a
	}

	return b
}

func IntAorB(a, b int) int {
	if a != 0 {
		return a
	}

	return b
}

func StripIllegalChars(name string) string {
	g2u := map[string]string{
		"/":  "_",
		"\\": "_",
		"\"": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	return strutil.Replaces(name, g2u)
}

func Filenamify(name string) string {
	g2u := map[string]string{
		"/":  "_",
		"\\": "_",
		"\"": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	return strutil.Replaces(name, g2u)
}
