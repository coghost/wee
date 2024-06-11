package wee

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-rod/rod/lib/proto"
	"github.com/gookit/goutil/fsutil"
	"github.com/rs/zerolog/log"
)

var cookieFolder = fsutil.Expand("~/tmp/wee_cookies")

func (b *Bot) DumpCookies() (string, error) {
	filepath, _ := b.uniqueNameByUrl(b.CurrentUrl())
	fsutil.MkParentDir(filepath)

	content, err := json.Marshal(b.page.MustCookies())
	if err != nil {
		return "", fmt.Errorf("cannot marshal cookies: %w", err)
	}

	var mode fs.FileMode = 0o644

	err = os.WriteFile(filepath, content, mode)
	if err != nil {
		return "", fmt.Errorf("cannot save file: %w", err)
	}

	return filepath, nil
}

func (b *Bot) LoadCookies(filepath string) error {
	if filepath == "" {
		v, err := b.uniqueNameByUrl(b.CurrentUrl())
		if err != nil {
			return err
		}

		filepath = v
	}

	if !fsutil.FileExist(filepath) {
		return ErrMissingCookieFile
	}

	raw, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var cookies []proto.NetworkCookie
	err = json.Unmarshal([]byte(raw), &cookies)
	b.pie(err)

	var nodes []*proto.NetworkCookieParam
	for _, cookie := range cookies {
		nodes = append(nodes, &proto.NetworkCookieParam{
			Name:         cookie.Name,
			Value:        cookie.Value,
			URL:          b.CurrentUrl(),
			Domain:       cookie.Domain,
			Path:         cookie.Path,
			Secure:       cookie.Secure,
			HTTPOnly:     cookie.HTTPOnly,
			SameSite:     cookie.SameSite,
			Expires:      cookie.Expires,
			Priority:     cookie.Priority,
			SameParty:    cookie.SameParty,
			SourceScheme: cookie.SourceScheme,
			SourcePort:   &cookie.SourcePort,
		})
	}

	err = b.page.SetCookies(nodes)
	if err != nil {
		log.Error().Err(err).Msg("load cookies failed")
	}

	return err
}

func (b *Bot) uniqueNameByUrl(uri string) (string, error) {
	if uri == "" {
		uri = b.CurrentUrl()
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.cookies", u.Scheme, u.Host)
	filepath := filepath.Join(cookieFolder, filename)

	return filepath, nil
}

func (b *Bot) FromCookieStr() {
}
