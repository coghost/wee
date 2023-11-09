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

	root *rod.Element

	// browserOnly is used when we init browser only.
	browserOnly bool
	userAgent   string

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
	bot := &Bot{}
	bot.Init()

	bindBotOptions(bot, options...)

	if !bot.browserOnly {
		bot.CustomizePage()
	}

	return bot
}

func NewBotWithDefault(options ...BotOption) *Bot {
	l, brw := NewBrowser()
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

func NewBotWithDebug(options ...BotOption) *Bot {
	l, brw := NewBrowser(WithPaintRects(true))
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

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

func WithBrowserOnly(b bool) BotOption {
	return func(o *Bot) {
		o.browserOnly = b
	}
}

func WithPanicBy(i PanicByType) BotOption {
	return func(o *Bot) {
		o.panicBy = i
	}
}

func WithPopovers(arr ...string) BotOption {
	return func(o *Bot) {
		o.popovers = append(o.popovers, arr...)
		o.popovers = lo.Uniq(o.popovers)
	}
}
