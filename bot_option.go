package wee

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/samber/lo"
)

type Bot struct {
	// init behaviours
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page

	userMode bool

	headless    bool
	withCookies bool

	cookieFile string

	root *rod.Element

	withPage  bool
	userAgent string

	highlightTimes int

	// popovers are anything that popup and block normal actions
	popovers []string

	//
	panicBy PanicByType

	longToSec   time.Duration
	mediumToSec time.Duration
	shortToSec  time.Duration
	napToSec    time.Duration
}

func NewBot(options ...BotOption) *Bot {
	bot := &Bot{withPage: true}
	bot.Init()

	bindBotOptions(bot, options...)

	if bot.headless {
		l, brw := NewBrowser(WithBrowserHeadless(bot.headless))
		bot.launcher = l
		bot.browser = brw
	}

	if bot.withPage {
		bot.CustomizePage()
	}

	return bot
}

func NewBotDefault(options ...BotOption) *Bot {
	l, brw := NewBrowser()
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

func NewBotDebug(options ...BotOption) *Bot {
	l, brw := NewBrowser(WithPaintRects(true))
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

func NewBotUserMode(options ...BotOption) *Bot {
	l, brw := NewUserMode()
	options = append(options, WithLauncher(l), WithBrowser(brw), setUserMode(true))

	return NewBot(options...)
}

/** bot init options **/

type BotOption func(*Bot)

func bindBotOptions(b *Bot, opts ...BotOption) {
	for _, f := range opts {
		f(b)
	}
}

func WithLauncher(l *launcher.Launcher) BotOption {
	return func(o *Bot) {
		o.launcher = l
	}
}

func WithBrowser(brw *rod.Browser) BotOption {
	return func(o *Bot) {
		o.browser = brw
	}
}

func WithUserAgent(s string) BotOption {
	return func(o *Bot) {
		o.userAgent = s
	}
}

func WithHighlightTimes(i int) BotOption {
	return func(o *Bot) {
		o.highlightTimes = i
	}
}

func WithPage(b bool) BotOption {
	return func(o *Bot) {
		o.withPage = b
	}
}

func WithCookies(b bool) BotOption {
	return func(o *Bot) {
		o.withCookies = b
	}
}

func WithCookieFile(s string) BotOption {
	return func(o *Bot) {
		o.cookieFile = s
	}
}

func Headless(b bool) BotOption {
	return func(o *Bot) {
		o.headless = b
	}
}

func setUserMode(b bool) BotOption {
	return func(o *Bot) {
		o.userMode = b
	}
}

func WithPanicBy(i PanicByType) BotOption {
	return func(o *Bot) {
		o.panicBy = i
	}
}

// WithPopovers is useful when popovers appear randomly.
func WithPopovers(popovers ...string) BotOption {
	return func(o *Bot) {
		o.popovers = append(o.popovers, popovers...)
		o.popovers = lo.Uniq(o.popovers)
	}
}
