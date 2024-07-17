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

var ErrEmptyCookieStr = errors.New("no cookie str found")

var cookieFolder = fsutil.Expand(".cookies")

func (b *Bot) DumpCookies() (string, error) {
	if b.cookieFile == "" {
		file, err := b.uniqueNameByURL(b.CurrentURL())
		if err != nil {
			return "", err
		}

		b.cookieFile = file
	}

	if err := fsutil.MkParentDir(b.cookieFile); err != nil {
		return "", err
	}

	content, err := json.Marshal(b.page.MustCookies())
	if err != nil {
		return "", fmt.Errorf("cannot marshal cookies: %w", err)
	}

	var mode fs.FileMode = 0o644

	err = os.WriteFile(b.cookieFile, content, mode)
	if err != nil {
		return "", fmt.Errorf("cannot save file: %w", err)
	}

	return b.cookieFile, nil
}

func (b *Bot) getRawCookies(filepath string) ([]byte, error) {
	// 1. check if copy as cURL
	if len(b.copyAsCURLCookies) != 0 {
		return b.copyAsCURLCookies, nil
	}

	// 2. filepath or auto get by URL
	if filepath == "" {
		v, err := b.uniqueNameByURL(b.CurrentURL())
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

func (b *Bot) LoadCookies(filepath string) ([]*proto.NetworkCookieParam, error) {
	raw, err := b.getRawCookies(filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot get raw cookies: %w", err)
	}

	if !IsJSON(string(raw)) {
		// non json format, means from cURL
		// parse as cURL
		return ParseCopyAsCURLAsCookie(string(raw))
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

func (b *Bot) uniqueNameByURL(uri string) (string, error) {
	if uri == "" {
		uri = b.CurrentURL()
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.cookies", u.Scheme, u.Host)
	folder := StrAorB(b.cookieFolder, cookieFolder)
	filepath := filepath.Join(folder, filename)

	return filepath, nil
}

func (b *Bot) GetCopyAsCURLCookie(filePath string) ([]*proto.NetworkCookieParam, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ParseCopyAsCURLAsCookie(string(raw))
}

func ParseCopyAsCURLAsCookie(raw string) ([]*proto.NetworkCookieParam, error) {
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
		port := 80

		if uriQuery.Scheme == "https" {
			httpOnly = false
			port = 443
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

func getHeader(header http.Header, key string) []string {
	cookieStr, b := header[key]
	if b {
		return cookieStr
	}

	return header[strings.ToLower(key)]
}

func flattenNodes(nodes []*proto.NetworkCookieParam) []string {
	got := []string{}
	for _, node := range nodes {
		got = append(got, fmt.Sprintf("%s=%s", node.Name, node.Value))
	}

	return got
}
