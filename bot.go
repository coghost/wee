package wee

import (
	"fmt"
	"time"

	"github.com/coghost/xpretty"
	"github.com/go-rod/rod"
	"github.com/gookit/goutil/dump"
	"github.com/rs/zerolog/log"
)

// Init init all attributes not exposed with options
func (b *Bot) Init() {
	b.highlightTimes = 1
	b.SetTimeout()
}

func (b *Bot) Cleanup() {
	if b.userMode {
		log.Debug().Msg("in user mode, no clean up")
		return
	}

	b.browser.MustClose()
	b.launcher.Cleanup()
}

func (b *Bot) CustomizePage() {
	if b.page == nil {
		b.page = b.browser.MustPage()
	}

	w, h := 1280, 728
	b.page = b.page.MustSetWindow(0, 0, w, h)
}

func (b *Bot) SetTimeout() {
	b.longToSec = LongToSec * time.Second
	b.mediumToSec = MediumToSec * time.Second
	b.shortToSec = ShortToSec * time.Second
	b.napToSec = NapToSec * time.Second
}

func (b *Bot) SetPanicWith(panicWith PanicByType) {
	b.panicBy = panicWith
}

func (b *Bot) pie(err error) {
	if err == nil {
		return
	}

	switch b.panicBy {
	case PanicByDump:
		dump.P(err)
		xpretty.DumpCallerStack()
	case PanicByLogError:
		log.Error().Err(err).Msg("error of bot")
	case PanicByLogFatal:
		log.Fatal().Err(err).Msg("error of bot")
	default:
		log.Panic().Err(err).Msg("error of bot")
	}
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

// EnsureUrlHas uses `decodeURIComponent(window.location.href).includes` to check if url has str
func (b *Bot) EnsureUrlHas(str string, timeouts ...float64) error {
	timeout := float64(MediumToSec)
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	}

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
