package wee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/gookit/goutil/fsutil"
)

var ErrEmptyCookieStr = errors.New("no cookie str found")

var cookieFolder = fsutil.Expand("cookies")

func (b *Bot) DumpCookies() (string, error) {
	cookieFile := b.cookieFile

	if cookieFile == "" {
		file, err := b.uniqueNameByURL(b.CurrentURL())
		if err != nil {
			return "", err
		}

		cookieFile = file
	}

	if err := fsutil.MkParentDir(cookieFile); err != nil {
		return "", err
	}

	content, err := json.Marshal(b.page.MustCookies())
	if err != nil {
		return "", fmt.Errorf("cannot marshal cookies: %w", err)
	}

	var mode fs.FileMode = 0o644

	err = os.WriteFile(cookieFile, content, mode)
	if err != nil {
		return "", fmt.Errorf("cannot save file: %w", err)
	}

	return cookieFile, nil
}

func (b *Bot) LoadCookies(filepath string) ([]*proto.NetworkCookieParam, error) {
	if raw := b.cURLFromClipboard; raw != "" {
		return ParseCopyAsCURLAsCookie(raw)
	}

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

	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// first try parse as cURL
	ckArr, err := ParseCopyAsCURLAsCookie(string(raw))
	if err != nil {
		return nil, err
	}

	if ckArr != nil {
		return ckArr, nil
	}

	var cookies []proto.NetworkCookie

	err = json.Unmarshal(raw, &cookies)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal to proto.cookie: %w", err)
	}

	var nodes []*proto.NetworkCookieParam

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
	req, b := parse(raw)
	if !b {
		return nil, nil
	}

	cookieStr, ok := req.Header["cookie"]
	if !ok {
		return nil, ErrEmptyCookieStr
	}

	uriQuery, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}

	// ckObjArr := []map[string]interface{}{}
	var nodes []*proto.NetworkCookieParam

	for _, pair := range strings.Split(cookieStr, "; ") {
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
