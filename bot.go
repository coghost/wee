package wee

import (
	"log"
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

const (
	_uniqueIDLen   = 16
	_browserChrome = "Google Chrome"
)

type Bot struct {
	// UniqueID is the identifier of a bot.
	UniqueID string

	logger *zap.Logger

	// trackTime tracks the time spend on operation.
	trackTime bool

	// init behaviours
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page
	prevPage *rod.Page
	root     *rod.Element

	left           int
	windowMaximize bool

	// bounds of browser
	bounds *BrowserBounds

	isLaunched       bool
	withPageCreation bool
	humanized        bool
	stealthMode      bool

	// launcher/browser options
	userMode       bool
	headless       bool
	userAgent      string
	acceptLanguage string
	userDataDir    string

	// when in userMode, by default will skip cleanup, we can set forceCleanup to do the cleanup.
	forceCleanup bool
	// clearCookies used when we want to clear all cookies in user-mode
	clearCookies bool
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
	pt1s          time.Duration

	browserOptions []BrowserOptionFunc
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

// NewBotWithOptionsOnly creates a new Bot instance with options only, without creating a page.
func NewBotWithOptionsOnly(options ...BotOption) *Bot {
	options = append(options, WithPage(false))
	return NewBot(options...)
}

// BindBotLanucher launches the browser and page for the bot.
// This is used when we create a bot first and launch the browser elsewhere.
func BindBotLanucher(bot *Bot, options ...BotOption) {
	if bot.isLaunched {
		return
	}

	var (
		lnchr *launcher.Launcher
		brw   *rod.Browser
	)

	if bot.userMode {
		if bot.forceCleanup {
			err := ForceQuitBrowser(_browserChrome, 5) //nolint: mnd
			if err != nil {
				log.Fatalf("cannot close chrome: %v", err)
			}
		}

		lnchr, brw = NewUserMode(LaunchLeakless(bot.forceCleanup), BrowserUserDataDir(bot.userDataDir))
		// lnchr, brw = NewUserMode(bot.browserOptions...)
	} else {
		lnchr, brw = NewBrowser(bot.browserOptions...)
	}

	options = append(options, Launcher(lnchr), Browser(brw), WithPage(true))
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

// NewBotDefault creates a bot with a default launcher/browser.
// The launcher/browser passed in will be ignored.
func NewBotDefault(options ...BotOption) *Bot {
	l, brw := NewBrowser()
	options = append(options, Launcher(l), Browser(brw))

	return NewBot(options...)
}

// NewBotHeadless sets headless with NewBotDefault
func NewBotHeadless(options ...BotOption) *Bot {
	options = append(options, Headless(true))
	return NewBotDefault(options...)
}

// NewBotForDebug creates a bot which launches browser with option: `--show-paint-rects`,
//
// refer: https://peter.sh/experiments/chromium-command-line-switches/#show-paint-rects
func NewBotForDebug(options ...BotOption) *Bot {
	l, brw := NewBrowser(BrowserPaintRects(true))
	options = append(options, Launcher(l), Browser(brw))

	return NewBot(options...)
}

// NewBotUserMode creates a bot that connects to the system Chrome browser.r
func NewBotUserMode(options ...BotOption) *Bot {
	l, brw := NewUserMode()
	options = append(options, Launcher(l), Browser(brw), UserMode(true))

	return NewBot(options...)
}

// initialize inits all attributes not exposed with options
func (b *Bot) initialize() {
	b.logger = zlog.MustNewZapLogger()

	b.highlightTimes = 1
	b.SetTimeout()
	b.UniqueID = strutil.RandomCharsV3(_uniqueIDLen)
}

// BlockInCleanUp performs cleanup and blocks execution.
func (b *Bot) BlockInCleanUp() {
	defer b.Cleanup()
	defer Blocked()
}

// Cleanup closes the opened page and quits the browser in non-userMode.
// In userMode, by default it will skip cleanup.
func (b *Bot) Cleanup() {
	if !b.isLaunched {
		return
	}

	// non user mode, close and clean.
	if !b.userMode {
		b.browser.MustClose()
		b.launcher.Cleanup()

		return
	}

	// by default is not force mode, just return.
	if !b.forceCleanup {
		b.logger.Info("runs in user mode, skip cleanup browser, you should call `bot.Page().Close()` manually.")
		return
	}

	// in case cookie matters the crawler result.
	if b.clearCookies {
		b.ClearStorageCookies()

		_ = RandSleepNap()
	}

	b.browser.MustClose()
	b.launcher.Cleanup()
}

// ClearStorageCookies calls StorageClearCookies, clears all history cookies.
func (b *Bot) ClearStorageCookies() {
	err := b.browser.SetCookies(nil)
	if err != nil {
		b.logger.Error("cannot clear cookies", zap.Error(err))
	} else {
		b.logger.Info("cleared cookies before quit.")
	}
}

func (b *Bot) SetTimeout() {
	b.longTimeout = LongToSec * time.Second
	b.mediumTimeout = MediumToSec * time.Second
	b.shortTimeout = ShortToSec * time.Second
	b.napTimeout = NapToSec * time.Second
	b.pt10s = PT10Sec * time.Second
	b.pt1s = 1 * time.Second
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

func (b *Bot) Page() *rod.Page {
	return b.page
}

func (b *Bot) Browser() *rod.Browser {
	return b.browser
}

func (b *Bot) CurrentURL() string {
	return b.page.MustInfo().URL
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

func UserDataDir(s string) BotOption {
	return func(o *Bot) {
		o.userDataDir = s
		o.browserOptions = append(o.browserOptions, BrowserUserDataDir(s))
	}
}

func WithBrowserOptions(browserOptions []BrowserOptionFunc) BotOption {
	return func(o *Bot) {
		o.browserOptions = append(o.browserOptions, browserOptions...)
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

func UserMode(b bool) BotOption {
	return func(o *Bot) {
		o.userMode = b
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

func ForceCleanup(b bool) BotOption {
	return func(o *Bot) {
		o.forceCleanup = b
	}
}

func ClearCookies(b bool) BotOption {
	return func(o *Bot) {
		o.clearCookies = b
	}
}

// TrackTime tracktime shows logs with level `debug`.
func TrackTime(b bool) BotOption {
	return func(o *Bot) {
		o.trackTime = b
	}
}

func Logger(l *zap.Logger) BotOption {
	return func(o *Bot) {
		o.logger = l
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

// WithLeftPosition sets the left position of the browser window.
func WithLeftPosition(i int) BotOption {
	return func(o *Bot) {
		o.left = i
	}
}
