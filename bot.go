package wee

import (
	"fmt"
	"time"

	"github.com/coghost/xpretty"
	"github.com/go-rod/rod"
	"github.com/gookit/goutil/dump"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Init init all attributes not exposed with options
func (b *Bot) Init() {
	b.highlightTimes = 1
	b.SetTimeout()
}

func (b *Bot) Cleanup() {
	b.browser.MustClose()
	b.launcher.Cleanup()
}

func (b *Bot) QuitOnTimeout(args ...int) {
	dft := 3

	sec := FirstOrDefault(dft, args...)
	if sec == 0 {
		return
	}

	if sec < 0 {
		Pause()
		return
	}

	SleepWithSpin(sec)
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

func (b *Bot) e(err error) {
	if err == nil {
		return
	}

	xpretty.DumpCallerStack()

	switch b.panicBy {
	case PanicByDump:
		dump.P(err)
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
		log.Fatal().Err(err).Msg("cannot open page")
	}

	// when page is loaded, set timeout to medium timeout
	b.page.Timeout(b.mediumToSec)
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

func (b *Bot) CurrentUrl() string {
	return b.page.MustInfo().URL
}

func (b *Bot) BindPopovers(arr ...string) {
	b.popovers = append(b.popovers, arr...)
	b.popovers = lo.Uniq(b.popovers)
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
