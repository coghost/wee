package wee

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
)

var ErrCannotActivateOpenedPage = errors.New("cannot activate latest opened page")

func (b *Bot) CustomizePage() {
	if b.page == nil {
		b.page = b.browser.MustPage()
	}

	ua := b.userAgent
	lang := b.acceptLanguage

	if ua != "" || lang != "" {
		ov := overrideUA(ua, lang)
		b.page.MustSetUserAgent(ov)
	}

	w, h := 1280, 768
	vw, vh := w, 728

	b.page = b.page.MustSetWindow(0, 0, w, h)
	b.page = b.page.MustSetViewport(vw, vh, 0.0, false)
}

func (b *Bot) MustOpen(url string, timeouts ...time.Duration) {
	if err := b.Open(url); err != nil {
		log.Error().Err(err).Msg("cannot open page")
		return
	}

	// when page is loaded, set timeout to medium timeout

	timeout := b.longToSec
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	}

	b.page.Timeout(timeout)

	if b.cookieFile != "" || b.withCookies {
		b.LoadCookies(b.cookieFile)
		b.page.Reload()
	}
}

func (b *Bot) Open(url string) error {
	if err := b.page.Timeout(b.longToSec).Navigate(url); err != nil {
		return err
	}

	return b.page.Timeout(b.longToSec).WaitLoad()
}

func (b *Bot) Page() *rod.Page {
	return b.page
}

func (b *Bot) Browser() *rod.Browser {
	return b.browser
}

func (b *Bot) CurrentUrl() string {
	return b.page.MustInfo().URL
}

// MustEval  a wrapper with MediumTo to rod.Page.MustEval
//
// refer: https://github.com/go-rod/rod/blob/main/examples_test.go#L53
//
//	@param script `() => console.log("hello world")` or "`(a, b) => a + b`, 1, 2"
//	@return string
func (b *Bot) MustEval(script string) string {
	res := b.page.Timeout(b.mediumToSec).MustEval(script).String()
	return res
}

// URLContains uses `decodeURIComponent(window.location.href).includes` to check if url has str or not
func (b *Bot) URLContains(str string, timeouts ...float64) error {
	timeout := FirstOrDefault(float64(MediumToSec), timeouts...)
	script := fmt.Sprintf(`() => decodeURIComponent(window.location.href).includes("%s")`, str)

	err := rod.Try(func() {
		b.page.Timeout(time.Second * time.Duration(timeout)).MustWait(script).CancelTimeout()
	})

	return err
}

func (b *Bot) PageSource() string {
	return b.page.MustHTML()
}

func (b *Bot) MustStable() {
	b.page.Timeout(b.mediumToSec).MustWaitStable().CancelTimeout()
}

func (b *Bot) MustWaitLoad() {
	b.page.Timeout(b.mediumToSec).MustWaitLoad().CancelTimeout()
}

func overrideUA(uaStr string, lang string) *proto.NetworkSetUserAgentOverride {
	uaOverride := proto.NetworkSetUserAgentOverride{}
	uaOverride.UserAgent = uaStr

	if lang == "" {
		lang = "en-CN,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,en-GB;q=0.6,en-US;q=0.5"
	}

	uaOverride.AcceptLanguage = lang

	return &uaOverride
}

func (b *Bot) UpdatePage(page *rod.Page) error {
	b.prevPage, b.page = b.page, page
	_, err := b.page.Activate()

	return err
}

func (b *Bot) MustUpdatePage(page *rod.Page) {
	err := b.UpdatePage(page)
	b.pie(err)
}

func (b *Bot) ResetToOriginalPage() error {
	if b.prevPage != nil && b.page != nil {
		err := b.page.Close()
		if err != nil {
			return fmt.Errorf("cannot reset original page by close page: %w", err)
		}
	}

	b.page = b.prevPage

	return nil
}

// ActivatePageByURLRegex
//
// usage:
//
//	if v := regexStr; v != "" {
//		return c.activatePageByURLRegex(v)
//	}
func (b *Bot) ActivatePageByURLRegex(jsRegex string, retry int) error {
	var page *rod.Page

	for i := 0; i < retry; i++ {
		page, _ = b.browser.MustPages().FindByURL(jsRegex)
		if page == nil {
			RandSleep(1, 1.1) //nolint:gomnd
			continue
		}

		break
	}

	return b.UpdatePage(page)
}

func (b *Bot) ActivateLatestOpenedPage(pagesBefore rod.Pages, retry int) error {
	var pageWant *rod.Page

	for i := 0; i < retry; i++ {
		pagesAfter := b.browser.MustPages()
		for _, np := range pagesAfter {
			if !funk.Contains(pagesBefore, np) {
				pageWant = np
				break
			}
		}

		if pageWant != nil {
			break
		}

		RandSleep(1.0, 1.1) //nolint: gomnd
	}

	if pageWant == nil {
		return ErrCannotActivateOpenedPage
	}

	return b.UpdatePage(pageWant)
}
