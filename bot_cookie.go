package wee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/gookit/goutil/fsutil"
	"github.com/jrefior/uncurl"
)

const (
	_defaultCookieFolder = ".cookies"
	// File permissions
	_cookieFilePermissions fs.FileMode = 0644

	// HTTP schemes
	_schemeHTTP  = "http"
	_schemeHTTPS = "https"

	// HTTP ports
	_portHTTP  = 80
	_portHTTPS = 443
)

var ErrEmptyCookieStr = errors.New("no cookie str found")

// DumpCookies saves the current cookies of the bot's page to a file on disk.
//
// This function performs the following operations:
//
// 1. Ensures the cookie file exists:
//   - If b.cookieFile is empty, it generates a filename based on the current URL.
//   - Creates any necessary parent directories for the cookie file.
//
// 2. Retrieves and marshals cookies:
//   - Calls b.page.MustCookies() to get all cookies from the current page.
//   - Marshals these cookies into JSON format.
//
// 3. Writes cookies to file:
//   - Opens or creates the cookie file with 0644 permissions (rw-r--r--).
//   - Writes the JSON-encoded cookie data to the file.
//
// Returns:
//   - string: The full path to the cookie file where the cookies were saved.
//   - error: An error if any of the following occur:
//   - Failure to ensure the cookie file (e.g., permission issues)
//   - Failure to marshal cookies to JSON
//   - Failure to write to the cookie file
//
// Usage:
//
//	filepath, err := bot.DumpCookies()
//	if err != nil {
//	    log.Fatalf("Failed to dump cookies: %v", err)
//	}
//	fmt.Printf("Cookies saved to: %s\n", filepath)
//
// Notes:
//   - This function is useful for persisting session data between bot runs.
//   - The saved cookies can be later loaded using the LoadCookies method.
//   - If b.cookieFile is not set, the function will automatically generate a filename
//     based on the current URL, stored in the default or specified cookie folder.
//   - Ensure the bot has necessary permissions to write to the cookie directory.
func (b *Bot) DumpCookies() (string, error) {
	err := b.ensureCookieFile()
	if err != nil {
		return "", err
	}

	content, err := json.Marshal(b.page.MustCookies())
	if err != nil {
		return "", fmt.Errorf("cannot marshal cookies: %w", err)
	}

	err = os.WriteFile(b.cookieFile, content, _cookieFilePermissions)
	if err != nil {
		return "", fmt.Errorf("cannot save file: %w", err)
	}

	return b.cookieFile, nil
}

// ensureCookieFile create cookie file if not existed.
func (b *Bot) ensureCookieFile() error {
	if b.cookieFile == "" {
		file, err := b.createCookieFilenameFromURL(b.CurrentURL())
		if err != nil {
			return err
		}

		b.cookieFile = file
	}

	if err := fsutil.MkParentDir(b.cookieFile); err != nil {
		return err
	}

	return nil
}

func (b *Bot) getRawCookies(filepath string) ([]byte, error) {
	// 1. check if copy as cURL
	if len(b.copyAsCURLCookies) != 0 {
		return b.copyAsCURLCookies, nil
	}

	// 2. filepath or auto get by URL
	if filepath == "" {
		v, err := b.createCookieFilenameFromURL(b.CurrentURL())
		if err != nil {
			return nil, err
		}

		filepath = v
	}

	if !fsutil.FileExist(filepath) {
		return nil, ErrMissingCookieFile
	}

	return os.ReadFile(filepath)
}

// LoadCookies loads cookies from a file or from previously stored cURL data and converts them
// into a format usable by the bot.
//
// Parameters:
//   - filepath: string - The path to the cookie file. If empty, the function will attempt to
//     generate a filename based on the bot's current URL.
//
// Returns:
//   - []*proto.NetworkCookieParam: A slice of NetworkCookieParam objects representing the loaded cookies.
//   - error: An error if the loading process fails at any step.
//
// Function behavior:
// 1. Retrieves raw cookie data:
//   - If copyAsCURLCookies is not empty, it uses this data.
//   - Otherwise, it reads from the specified file or an auto-generated file based on the current URL.
//
// 2. Determines the format of the cookie data:
//   - If the data is in JSON format, it parses it as proto.NetworkCookie objects.
//   - If not JSON, it assumes cURL format and parses accordingly.
//
// 3. For JSON format:
//   - Unmarshals the data into proto.NetworkCookie objects.
//   - Converts these into proto.NetworkCookieParam objects, filling in additional fields.
//
// 4. For cURL format:
//   - Calls ParseCURLCookies to handle the conversion.
//
// Usage:
//
//	cookies, err := bot.LoadCookies("path/to/cookies.json")
//	if err != nil {
//	    log.Fatalf("Failed to load cookies: %v", err)
//	}
//	// Use the cookies with the bot...
//
// Notes:
//   - This function is flexible and can handle both JSON-formatted cookie files (typically saved
//     by DumpCookies) and cURL-formatted cookie strings.
//   - When loading from a file, ensure the bot has read permissions for the cookie file.
//   - If filepath is empty and the bot can't determine the current URL, this will result in an error.
//   - The returned cookies are in a format ready to be used with the bot's page or browser instance.
//   - Any parsing or file reading errors will be returned, allowing the caller to handle them appropriately.
//
// See also:
// - DumpCookies: For saving cookies to a file.
// - ParseCURLCookies: For the specific handling of cURL-formatted cookie strings.
func (b *Bot) LoadCookies(filepath string) ([]*proto.NetworkCookieParam, error) {
	raw, err := b.getRawCookies(filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot get raw cookies: %w", err)
	}

	if !IsJSON(string(raw)) {
		// non json format, means from cURL
		// parse as cURL
		return ParseCURLCookies(string(raw))
	}

	// if not copyAsCURL cookies, try parse with json format.
	var (
		cookies []proto.NetworkCookie
		nodes   []*proto.NetworkCookieParam
	)

	if err := json.Unmarshal(raw, &cookies); err != nil {
		return nil, fmt.Errorf("cannot unmarshal to proto.cookie: %w", err)
	}

	for _, cookie := range cookies {
		port := &cookie.SourcePort
		nodes = append(nodes, &proto.NetworkCookieParam{
			Name:         cookie.Name,
			Value:        cookie.Value,
			URL:          b.CurrentURL(),
			Domain:       cookie.Domain,
			Path:         cookie.Path,
			Secure:       cookie.Secure,
			HTTPOnly:     cookie.HTTPOnly,
			SameSite:     cookie.SameSite,
			Expires:      cookie.Expires,
			Priority:     cookie.Priority,
			SameParty:    cookie.SameParty,
			SourceScheme: cookie.SourceScheme,
			SourcePort:   port,
		})
	}

	return nodes, nil
}

// createCookieFilenameFromURL generates a cookie filename based on the given URL.
// If no URL is provided, it uses the bot's current URL.
// Returns the full filepath for the cookie file and any error encountered.
func (b *Bot) createCookieFilenameFromURL(uri string) (string, error) {
	if uri == "" {
		uri = b.CurrentURL()
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.cookies", u.Scheme, u.Host)
	folder := StrAorB(b.cookieFolder, _defaultCookieFolder)
	filepath := filepath.Join(folder, filename)

	return filepath, nil
}

// ParseCURLCookiesFromFile reads a file containing cURL-formatted cookies,
// parses the content, and returns NetworkCookieParam objects.
// It returns an error if the file cannot be read or parsed.
func (b *Bot) ParseCURLCookiesFromFile(filePath string) ([]*proto.NetworkCookieParam, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ParseCURLCookies(string(raw))
}

// ParseCURLCookies converts a cURL-formatted cookie string into NetworkCookieParam objects.
//
// It extracts cookies from the "Cookie" header in the cURL string, parses each cookie,
// and creates NetworkCookieParam objects with appropriate attributes based on the URL.
//
// Parameters:
//   - raw: string - The raw cURL-formatted cookie string.
//
// Returns:
//   - []*proto.NetworkCookieParam: Slice of parsed cookie objects.
//   - error: Error if parsing fails or no cookies are found.
//
// Usage:
//
//	cookies, err := ParseCURLCookies(curlString)
func ParseCURLCookies(raw string) ([]*proto.NetworkCookieParam, error) {
	curl, err := uncurl.NewString(raw)
	if err != nil {
		return nil, fmt.Errorf("cannot new copy as curl: %w", err)
	}

	cookieArr := getHeader(curl.Header(), "Cookie")
	if len(cookieArr) == 0 {
		return nil, ErrEmptyCookieStr
	}

	uriQuery, err := url.Parse(curl.Target())
	if err != nil {
		return nil, err
	}

	var nodes []*proto.NetworkCookieParam

	for _, pair := range strings.Split(cookieArr[0], "; ") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		arr := strings.Split(pair, "=")
		httpOnly := true
		port := _portHTTP

		if uriQuery.Scheme == _schemeHTTPS {
			httpOnly = false
			port = _portHTTPS
		}

		ckObj := &proto.NetworkCookieParam{
			Name:       arr[0],
			Value:      arr[1],
			Domain:     uriQuery.Host,
			Path:       uriQuery.Path,
			HTTPOnly:   httpOnly,
			SourcePort: &port,
		}

		nodes = append(nodes, ckObj)
	}

	return nodes, nil
}

// getHeader retrieves the values for a given header key from an http.Header object.
// It checks for both the original key and its lowercase version.
// Returns a slice of string values associated with the header key.
func getHeader(header http.Header, key string) []string {
	cookieStr, b := header[key]
	if b {
		return cookieStr
	}

	return header[strings.ToLower(key)]
}

// flattenNodes converts a slice of NetworkCookieParam objects into a slice of strings,
// where each string is in the format "name=value".
// Returns a slice of strings representing the flattened cookie data.
func flattenNodes(nodes []*proto.NetworkCookieParam) []string {
	got := []string{}
	for _, node := range nodes {
		got = append(got, fmt.Sprintf("%s=%s", node.Name, node.Value))
	}

	return got
}
