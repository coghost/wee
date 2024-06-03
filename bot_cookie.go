package wee

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

func (b *Bot) DumpCookies() string {
	filepath, _ := b.uniqueNameByUrl(b.CurrentUrl())

	content, err := json.Marshal(b.page.MustCookies())
	b.pie(err)

	var mode fs.FileMode = 0o644
	err = os.WriteFile(filepath, content, mode)
	b.pie(err)

	return filepath
}

func (b *Bot) LoadCookies(filepath string) error {
	if filepath == "" {
		v, err := b.uniqueNameByUrl(b.CurrentUrl())
		if err != nil {
			return err
		}

		filepath = v
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
	filepath := filepath.Join("/tmp", filename)

	return filepath, nil
}

func (b *Bot) fromCookieStr() {
}
