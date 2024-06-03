package wee

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
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
	rand.NewSource(time.Now().Unix())

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

// NewChromeExtension will create two files from line, and save to savePath
//   - background.js
//   - manifest.json
//
// line format is: "host:port:username:password:<OTHER>"
func NewChromeExtension(line, savePath string) (string, string, string) {
	proxy_js := `var config = {
  mode: 'fixed_servers',
  rules: {
    singleProxy: {
      scheme: 'http',
      host: '%s',
      port: parseInt(%s),
    },
    bypassList: ['foobar.com'],
  },
}
chrome.proxy.settings.set({ value: config, scope: 'regular' }, function () {})
function callbackFn(details) {
  return {
    authCredentials: {
      username: '%s',
      password: '%s',
    },
  }
}
chrome.webRequest.onAuthRequired.addListener(callbackFn, { urls: ['<all_urls>'] }, ['blocking'])`
	manifest := `{
    "version": "1.0.0",
    "manifest_version": 2,
    "name": "GoccProxy",
    "permissions": ["proxy", "tabs", "unlimitedStorage", "storage", "<all_urls>", "webRequest", "webRequestBlocking"],
    "background": {
        "scripts": ["background.js"]
    },
    "minimum_chrome_version": "22.0.0"
}`
	arr := strings.Split(line, ":")
	proxy_js = fmt.Sprintf(proxy_js, arr[0], arr[1], arr[2], arr[3])

	ipAddr := arr[0]

	if strings.Contains(arr[0], "superproxy") {
		ipArr := strings.Split(arr[2], "-")
		ipAddr = ipArr[len(ipArr)-1]
	}

	ipAddr = strings.ReplaceAll(ipAddr, ".", "_")

	baseDir := filepath.Join(savePath, ipAddr)

	bg := filepath.Join(baseDir, "background.js")
	mf := filepath.Join(baseDir, "manifest.json")

	MustWriteFile(bg, []byte(proxy_js))
	MustWriteFile(mf, []byte(manifest))

	return baseDir, bg, mf
}

func MustWriteFile(filepath string, content []byte) {
	var mode fs.FileMode = 0o644

	err := os.WriteFile(filepath, content, mode)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot write to file")
	}
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
