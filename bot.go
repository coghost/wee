package wee

import (
	"time"

	"github.com/coghost/xpretty"
	"github.com/coghost/zlog"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/strutil"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Bot struct {
	// UniqueID is the identifier of a bot.
	UniqueID string

	logger *zap.Logger

	// init behaviours
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page
	prevPage *rod.Page
	root     *rod.Element

	isLaunched       bool
	withPageCreation bool
	humanized        bool
	stealthMode      bool

	// launcher/browser options
	userMode       bool
	headless       bool
	userAgent      string
	acceptLanguage string
	extensions     []string

	// cookies
	withCookies  bool
	cookieFolder string
	cookieFile   string
	// copyAsCURLCookies reads `copy as cURL` directly from clipboard
	copyAsCURLCookies []byte

	// scrollHeight is the last valid height, use as fallback value.
	scrollHeight float64

	// highlightTimes highlight element times before next operation.
	highlightTimes int

	// popovers are anything that popup and block normal actions,
	// if popovers is set, will try to interact with it every time before doing input/click operations.
	popovers []string

	panicBy PanicByType

	longTimeout   time.Duration
	mediumTimeout time.Duration
	shortTimeout  time.Duration
	napTimeout    time.Duration
	pt10s         time.Duration

	// unused options
	// innerHeight  int
	// presetOptions []BotOption
	// windowSize []int
	// viewSize   []int
	// incognito bool
}

func NewBot(options ...BotOption) *Bot {
	bot := &Bot{withPageCreation: true}
	bot.initialize()

	bindBotOptions(bot, options...)

	resetIfHeadless(bot)
	bot.CustomizePage()

	return bot
}

func NewBotWithOptionsOnly(options ...BotOption) *Bot {
	options = append(options, WithPage(false))
	return NewBot(options...)
}

// BindBotLanucher launches browser and page for bot.
// this is used when we create bot first, and launch browser at somewhere else.
func BindBotLanucher(bot *Bot, options ...BotOption) {
	if bot.isLaunched {
		return
	}

	l, brw := NewBrowser()
	options = append(options, Launcher(l), Browser(brw), WithPage(true))
	bindBotOptions(bot, options...)

	resetIfHeadless(bot)
	bot.CustomizePage()
}

func resetIfHeadless(bot *Bot) {
	if bot.headless {
		l, brw := NewBrowser(BrowserHeadless(bot.headless))
		bot.launcher = l
		bot.browser = brw
	}
}

// NewBotDefault creates a bot with a default launcher/browser,
// the launcher/browser passed in will be ignored.
func NewBotDefault(options ...BotOption) *Bot {
	l, brw := NewBrowser()
	options = append(options, Launcher(l), Browser(brw))

	return NewBot(options...)
}

// NewBotForDebug creates a bot which launches browser with option: `--show-paint-rects`,
//
// refer: https://peter.sh/experiments/chromium-command-line-switches/#show-paint-rects
func NewBotForDebug(options ...BotOption) *Bot {
	l, brw := NewBrowser(BrowserPaintRects(true))
	options = append(options, Launcher(l), Browser(brw))

	return NewBot(options...)
}

// NewBotUserMode calls to go-rod NewUserMode, connects to system chrome browser
func NewBotUserMode(options ...BotOption) *Bot {
	l, brw := NewUserMode()
	options = append(options, Launcher(l), Browser(brw), setUserMode(true))

	return NewBot(options...)
}

// initialize inits all attributes not exposed with options
func (b *Bot) initialize() {
	b.logger = zlog.MustNewZapLogger()

	b.highlightTimes = 1
	b.SetTimeout()
	b.UniqueID = strutil.RandomCharsV3(16) //nolint:mnd
}

func (b *Bot) BlockInCleanUp() {
	defer b.Cleanup()
	defer Blocked()
}

func (b *Bot) Cleanup() {
	if b.userMode {
		b.logger.Debug("runs with user mode, skip cleanup")
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
		b.logger.Error("error happens", zap.Error(err))
	default:
		dump.P(err)
		b.logger.Panic("error happens", zap.Error(err))
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

func Launcher(l *launcher.Launcher) BotOption {
	return func(o *Bot) {
		o.launcher = l
	}
}

func Browser(brw *rod.Browser) BotOption {
	return func(o *Bot) {
		o.browser = brw
	}
}

func UserAgent(s string) BotOption {
	return func(o *Bot) {
		o.userAgent = s
	}
}

func AcceptLanguage(s string) BotOption {
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
		o.withPageCreation = b
	}
}

// WithCookies will automatically load cookies from current dir for ".cookies/xxx".
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

// CopyAsCURLCookies reads the raw content `Copy as cURL` from clipboard
func CopyAsCURLCookies(b []byte) BotOption {
	return func(o *Bot) {
		o.copyAsCURLCookies = b
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

func Humanized(b bool) BotOption {
	return func(o *Bot) {
		o.humanized = b
	}
}

func StealthMode(b bool) BotOption {
	return func(o *Bot) {
		o.stealthMode = b
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
