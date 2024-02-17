package wee

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func (b *Bot) CustomizePage() {
	if b.page == nil {
		b.page = b.browser.MustPage()
	}

	w, h := 1280, 728
	b.page = b.page.MustSetWindow(0, 0, w, h)
}

func (b *Bot) MustOpen(url string) {
	if err := b.Open(url); err != nil {
		log.Error().Err(err).Msg("cannot open page")
		return
	}

	// when page is loaded, set timeout to medium timeout
	b.page.Timeout(b.mediumToSec)

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
