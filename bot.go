package wee

import (
	"time"

	"github.com/coghost/xlog"
	"github.com/coghost/xpretty"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/strutil"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type Bot struct {
	UniqueID string
	// init behaviours
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page
	prevPage *rod.Page

	isLaunched bool

	userMode bool

	headless bool

	withCookies  bool
	cookieFolder string
	cookieFile   string
	// cURLFromClipboard
	cURLFromClipboard string

	extensions []string

	root *rod.Element

	withPage  bool
	incognito bool

	userAgent      string
	acceptLanguage string

	windowSize []int
	viewSize   []int

	scrollHeight float64
	innerHeight  int

	highlightTimes int

	// popovers are anything that popup and block normal actions
	popovers []string

	//
	panicBy PanicByType

	presetOptions []BotOption

	longTimeout   time.Duration
	mediumTimeout time.Duration
	shortTimeout  time.Duration
	napTimeout    time.Duration
	pt10s         time.Duration
}

func NewBot(options ...BotOption) *Bot {
	bot := &Bot{withPage: true}
	bot.Init()

	bindBotOptions(bot, options...)

	checkIfHeadless(bot)
	bot.CustomizePage()

	return bot
}

func NewBotWithOptionsOnly(options ...BotOption) *Bot {
	options = append(options, WithPage(false))
	return NewBot(options...)
}

// BindBotLanucher launches browser and page for bot.
func BindBotLanucher(bot *Bot, options ...BotOption) {
	if bot.isLaunched {
		log.Debug().Msg("browser is already launched")
		return
	}

	l, brw := NewBrowser()
	options = append(options, WithLauncher(l), WithBrowser(brw), WithPage(true))
	bindBotOptions(bot, options...)

	checkIfHeadless(bot)
	bot.CustomizePage()
}

func checkIfHeadless(bot *Bot) {
	if bot.headless {
		l, brw := NewBrowser(WithBrowserHeadless(bot.headless))
		bot.launcher = l
		bot.browser = brw
	}
}

func NewBotDefault(options ...BotOption) *Bot {
	l, brw := NewBrowser()
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

func NewBotForDebug(options ...BotOption) *Bot {
	l, brw := NewBrowser(WithPaintRects(true))
	options = append(options, WithLauncher(l), WithBrowser(brw))

	return NewBot(options...)
}

func NewBotUserMode(options ...BotOption) *Bot {
	l, brw := NewUserMode()
	options = append(options, WithLauncher(l), WithBrowser(brw), setUserMode(true))

	return NewBot(options...)
}

// Init init all attributes not exposed with options
func (b *Bot) Init() {
	b.highlightTimes = 1
	b.SetTimeout()
	b.UniqueID = strutil.RandomCharsV3(16) //nolint:mnd
	xlog.InitLogDebug()
}

func (b *Bot) BlockInCleanUp() {
	defer b.Cleanup()
	defer Blocked()
}

func (b *Bot) Cleanup() {
	if b.userMode {
		// b.page.Close()
		log.Debug().Msg("in user mode, no clean up")

		return
	}

	b.browser.MustClose()
	b.launcher.Cleanup()
}

func (b *Bot) SetTimeout() {
	b.longTimeout = LongToSec * time.Second
	b.mediumTimeout = MediumToSec * time.Second
	b.shortTimeout = ShortToSec * time.Second
	b.napTimeout = NapToSec * time.Second
	b.pt10s = PT10Sec * time.Second
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
	default:
		dump.P(err)
		log.Panic().Err(err).Msg("error of bot")
	}
}

func (b *Bot) CookieFile() string {
	return b.cookieFile
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

func WithAcceptLanguage(s string) BotOption {
	return func(o *Bot) {
		o.acceptLanguage = s
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

// WithCookies will automatically load cookies from current dir for "cookies/xxx".
func WithCookies(b bool) BotOption {
	return func(o *Bot) {
		o.withCookies = b
	}
}

// WithCookieFolder will load cookies from folder passed in.
func WithCookieFolder(s string) BotOption {
	return func(o *Bot) {
		o.cookieFolder = s
	}
}

// WithCookieFile uses the cookie file specified,
// cookie file should be json array.
func WithCookieFile(s string) BotOption {
	return func(o *Bot) {
		o.cookieFile = s
	}
}

// WithCURLFromClipboard reads the raw content `Copy as cURL` from clipboard
func WithCURLFromClipboard(s string) BotOption {
	return func(o *Bot) {
		o.cURLFromClipboard = s
	}
}

func WithExtensionFolder(arr []string) BotOption {
	return func(o *Bot) {
		o.extensions = arr
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
