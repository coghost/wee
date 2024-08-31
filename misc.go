package wee

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/gookit/goutil/strutil"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cast"
)

// FirstOrDefault returns the first value from the variadic args if provided, otherwise returns the default value.
// It uses a generic type T to work with any data type.
//
// Parameters:
//   - dft: The default value to return if no args are provided.
//   - args: A variadic parameter of type T.
//
// Returns:
//   - The first value from args if provided, otherwise the default value dft.
func FirstOrDefault[T any](dft T, args ...T) T { //nolint:ireturn
	val := dft
	if len(args) > 0 {
		val = args[0]
	}

	return val
}

// NewStringSlice splits a raw string into a slice of strings.
//
// Parameters:
//   - raw: The input string to be split into a slice.
//   - fixStep: The base number of characters for each substring.
//   - randomStep: Optional boolean to enable random step mode. If true, the step size will vary randomly.
//
// Returns:
//   - A slice of strings, where each string has approximately fixStep characters.
//
// The function splits the input string into substrings. If randomStep is enabled,
// the actual number of characters in each substring may vary slightly from fixStep.
// This variation is determined by random chance, potentially adding or subtracting
// 1 or 2 characters from fixStep.
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

// NormalizeSliceIndex adjusts an index to be within the bounds of a slice or array.
//
// Parameters:
//   - length: The length of the slice or array.
//   - index: The input index to be normalized.
//
// Returns:
//   - An integer representing the normalized index.
//
// If length is 0, it returns -1.
// For negative indices, it wraps around from the end (e.g., -1 becomes length-1).
// If the index is out of bounds, it normalizes to the valid range [0, length-1].
func NormalizeSliceIndex(length, index int) int {
	if length == 0 {
		return -1
	}

	if index < 0 {
		index = length + index
	}

	index = min(index, length-1)

	return index
}

// Stringify converts an interface{} to its JSON string representation.
//
// Parameters:
//   - data: The interface{} to be converted to a JSON string.
//
// Returns:
//   - A string containing the JSON representation of the input data.
//   - An error if the conversion to JSON fails.
//
// This function uses json.Marshal to convert the input data to a JSON byte slice,
// then converts the byte slice to a string. If the marshaling process fails,
// it returns an empty string and the error from json.Marshal.
func Stringify(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// MustStringify converts an interface{} to its JSON string representation.
// It's similar to Stringify but panics if an error occurs during the conversion.
//
// Parameters:
//   - data: The interface{} to be converted to a JSON string.
//
// Returns:
//   - A string containing the JSON representation of the input data.
//
// This function calls Stringify internally and panics if an error is returned.
// It should be used only when you're certain that the conversion will succeed,
// or if you want to halt execution on failure.
func MustStringify(data interface{}) string {
	raw, err := Stringify(data)
	if err != nil {
		panic(err)
	}

	return raw
}

// MustStrToFloat converts a string to a float64, keeping only specified characters.
// It panics if the conversion fails.
//
// Parameters:
//   - raw: The input string to be converted.
//   - keptChars: A string containing characters to be kept along with digits.
//
// Returns:
//   - A float64 representation of the filtered string.
//
// This function first calls StrToNumChars to filter the input string,
// then uses cast.ToFloat64 to convert the result to a float64.
// If the conversion fails, it will panic.
func MustStrToFloat(raw string, keptChars string) float64 {
	v := StrToNumChars(raw, keptChars)
	return cast.ToFloat64(v)
}

// StrToNumChars extracts numeric characters and specified kept characters from a string.
//
// Parameters:
//   - raw: The input string to be processed.
//   - keptChars: A string containing additional characters to be kept along with digits.
//
// Returns:
//   - A string containing only digits and the specified kept characters.
//
// This function creates a regular expression pattern to match digits and the specified
// kept characters, finds all matching substrings in the input, and joins them together.
func StrToNumChars(raw string, keptChars string) string {
	chars := "[0-9" + keptChars + "]+"
	re := regexp.MustCompile(chars)
	c := re.FindAllString(raw, -1)
	r := strings.Join(c, "")

	return r
}

// IsJSON checks if a string is valid JSON.
//
// Parameters:
//   - str: The input string to be checked.
//
// Returns:
//   - A boolean indicating whether the input string is valid JSON.
//
// This function attempts to unmarshal the input string into a json.RawMessage.
// If the unmarshal operation succeeds (returns nil error), the function returns true,
// indicating that the input is valid JSON. Otherwise, it returns false.
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsZeroVal checks if a value of any type is its zero value.
//
// Parameters:
//   - x: The interface{} to be checked.
//
// Returns:
//   - A boolean indicating whether x is its zero value.
//
// This function checks if x is nil or if it's equal to its type's zero value.
// It uses reflect.DeepEqual to compare x with the zero value of its type.
// The function works for all types, including primitives, structs, and pointers.
func IsZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// StrAorB returns the first non-empty string between two input strings.
//
// Parameters:
//   - a: The first string to check.
//   - b: The second string to check.
//
// Returns:
//   - If 'a' is not empty, it returns 'a'.
//   - If 'a' is empty, it returns 'b', regardless of whether 'b' is empty or not.
//
// This function is useful for providing a fallback value when the primary string might be empty.
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

func FloatAorB(a, b float64) float64 {
	if a != 0 {
		return a
	}

	return b
}

// Filenamify converts a string into a valid filename by replacing illegal characters.
//
// Parameters:
//   - name: The input string to be converted into a valid filename.
//
// Returns:
//   - A string with illegal filename characters replaced by underscores.
//
// This function replaces common illegal filename characters (such as /, \, ", :, *, ?, <, >, |)
// with underscores. It's useful for sanitizing strings that will be used as filenames,
// ensuring they are compatible with most file systems. The function uses strutil.Replaces
// to perform the replacements based on a predefined map of illegal characters to underscores.
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

// NameFromURL extracts a filename-safe string from a URL, excluding the domain.
//
// Parameters:
//   - uri: The input URL string to be processed.
//
// Returns:
//   - A string derived from the URL path, suitable for use as a filename.
//
// This function attempts to parse the input URL and extract a meaningful name from its path.
// It removes the scheme and domain, trims leading and trailing slashes, and handles empty paths.
// If the resulting name is empty, it returns "homepage". The final result is passed through
// the Filenamify function to ensure it's safe for use as a filename.
//
// Example:
//
//	https://en.wikipedia.org/wiki/Main_Page ==> wiki_Main_Page
//
// If URL parsing fails, the function returns the original input string.
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

// NoEcho is a no-op function that accepts any number of arguments and does nothing.
//
// This function can be used as a placeholder or to explicitly ignore values.
// It's particularly useful in situations where you need to satisfy an interface
// that requires a function, but you don't actually need to do anything with the arguments.
//
// Parameters:
//   - ...any: Variadic parameter that can accept any type and any number of arguments.
//
// Returns:
//   - Nothing.
func NoEcho(...any) {}

// RetryIn3 attempts to execute a retryable function up to 3 times with a 1-second delay between attempts.
//
// Parameters:
//   - retryableFunc: A function of type retry.RetryableFunc to be executed with retry logic.
//
// Returns:
//   - An error if all retry attempts fail, or nil if the function succeeds within 3 attempts.
//
// This function uses the retry.Do method with the following options:
//   - LastErrorOnly(true): Only the last error is returned if all attempts fail.
//   - Attempts(3): The function will be attempted a maximum of 3 times.
//   - Delay(time.Second*1): There's a 1-second delay between each attempt.
func RetryIn3(retryableFunc retry.RetryableFunc) error {
	return retry.Do(
		retryableFunc,
		retry.LastErrorOnly(true),
		retry.Attempts(3), //nolint:mnd
		retry.Delay(time.Second*1))
}

// ForceQuitBrowser attempts to close a browser by killing its process.
//
// Parameters:
//   - browserName: The name of the browser process to be killed.
//   - retry: The number of times to retry killing the process if unsuccessful.
//
// Returns:
//   - An error if the process couldn't be killed after all retry attempts, nil otherwise.
//
// This function repeatedly tries to kill the specified browser process using the KillProcess function.
// If unsuccessful, it will retry up to the specified number of times, with a random sleep interval
// between attempts (using RandSleepNap). If all attempts fail, it returns the last error encountered.
func ForceQuitBrowser(browserName string, retry int) error {
	for i := 0; ; i++ {
		// pkill -a -i "Google Chrome"
		err := KillProcess(browserName)
		if err == nil {
			break
		}

		if i >= retry {
			return err
		}

		RandSleepNap()
	}

	return nil
}

// KillProcess terminates a process with the given name.
//
// Parameters:
//   - name: The name of the process to be terminated.
//
// Returns:
//   - An error if the process list couldn't be retrieved or if the process couldn't be killed.
//     Returns nil if the process was successfully terminated or if no matching process was found.
//
// This function iterates through all running processes, compares their names with the provided name,
// and attempts to kill the first matching process. If multiple processes have the same name,
// only the first encountered will be terminated. The function stops searching after killing one process.
func KillProcess(name string) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	for _, proc := range processes {
		n, _ := proc.Name()
		if n == "" {
			continue
		}

		if n == name {
			e := proc.Kill()
			if e != nil {
				return e
			}
		}
	}

	return nil
}
