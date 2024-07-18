package wee

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go"
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
	if min == max {
		return int(secToMill * min)
	}

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
	raw, err := Stringify(data)
	if err != nil {
		panic(err)
	}

	return raw
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

// NameFromURL gets a name from url without domain
//
//	i.e. https://en.wikipedia.org/wiki/Main_Page ==> wiki_Main_Page
func NameFromURL(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	name := uri
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")

	homepage := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	name = strings.ReplaceAll(name, homepage, "")

	if name == "" {
		return "homepage"
	}

	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")

	if name == "" {
		name = "homepage"
	}

	return Filenamify(name)
}

// NoEcho eats everything.
func NoEcho(...any) {}

func RetryIn3(retryableFunc retry.RetryableFunc) error {
	return retry.Do(
		retryableFunc,
		retry.LastErrorOnly(true),
		retry.Attempts(3), //nolint:mnd
		retry.Delay(time.Second*1))
}
