package wee

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coghost/wee/fixtures"
	"github.com/coghost/zlog"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/remeh/sizedwaitgroup"
	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BotSuite struct {
	suite.Suite
	ts *httptest.Server
}

const (
	_ip42pl = `https://ip.42.pl/`

	_mitmHTTPHome  = `http://mitm.it/`
	_mitmHTTPSHome = `https://mitmproxy.org/`

	_testExtension = "fixtures/chrome-extension"
)

func TestBot(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

func (s *BotSuite) SetupSuite() {
	s.T().Parallel()
	s.ts = fixtures.NewTestServer()
}

func (s *BotSuite) TearDownSuite() {
}

func (s *BotSuite) Test00NewBot() {
	s.T().Parallel()

	want := "<html><head></head>\n<body>\nhellowee\n\n\n</body></html>"
	l, brw := NewBrowser()

	tests := []struct {
		name  string
		bot   *Bot
		wantS string
	}{
		{
			name:  "new_bot",
			bot:   NewBot(Headless(true), Launcher(l), Browser(brw)),
			wantS: want,
		},
		{
			name:  "default_bot",
			bot:   NewBotDefault(),
			wantS: want,
		},
		{
			name:  "debug_bot",
			bot:   NewBotForDebug(),
			wantS: want,
		},
		{
			name:  "headless_bot",
			bot:   NewBotHeadless(),
			wantS: want,
		},
	}

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.bot.Cleanup()
			tt.bot.MustOpen(s.ts.URL)

			raw := tt.bot.page.MustHTML()
			raw = strings.ReplaceAll(raw, " ", "")

			s.Equal(tt.wantS, raw, "")
		})
	}
}

func (s *BotSuite) Test00BotWithOptionsOnly() {
	s.T().Parallel()

	bot := NewBotWithOptionsOnly(Headless(true), WithHighlightTimes(3), WithPanicBy(PanicByDft))
	defer bot.Cleanup()

	s.Nil(bot.page)
	s.Panics(func() {
		bot.MustOpen(s.ts.URL)
	}, "new bot only with options panics when open url")

	BindBotLanucher(bot)

	s.NotNil(bot.page)
	s.NotPanics(func() {
		bot.MustOpen(s.ts.URL)
		bot.MustWaitLoad()
	}, "after binding bot with browser, shouldn't panic")

	// another try to bind bot again, does nothing.
	BindBotLanucher(bot)

	bot.SetPanicWith(PanicByDump)
	// when panic by dump, pie will just print the error.
	bot.MustElem(`div.should-not-existed`, WithTimeout(PT1Sec))

	bot.SetPanicWith(PanicByLogError)
	// only log as error
	bot.MustElem(`div.should-not-existed`, WithTimeout(PT1Sec))

	s.NotNil(bot.Browser())
	s.Empty(bot.CookieFile())

	s.Equal(3, bot.highlightTimes)
}

func (s *BotSuite) Test00BotUserMode() {
	s.T().Parallel()

	bot := NewBotUserMode()
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL)
}

func (s *BotSuite) Test00BotStealth() {
	s.T().Parallel()

	sanny := `https://bot.sannysoft.com`
	sannyLoc := `td#chrome-result`

	tests := []struct {
		name    string
		stealth bool
		home    string
		loc     string
		wantS   string
	}{
		{
			name:  "sanny-headless",
			home:  sanny,
			loc:   sannyLoc,
			wantS: `missing (failed)`,
		},
		{
			name:    "sanny-stealth",
			home:    sanny,
			loc:     sannyLoc,
			stealth: true,
			wantS:   `present (passed)`,
		},
	}

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			bot := NewBotHeadless(StealthMode(tt.stealth))
			bot.MustOpen(tt.home)

			got := bot.MustElemAttr(tt.loc)
			s.Equal(tt.wantS, got, tt.name)
		})
	}
}

func (s *BotSuite) Test01CustomPageWithUAAndLang() {
	s.T().Parallel()

	userAgent := `wee - https://github.com/coghost/wee`
	lang := `en-CN,en`
	headers := []string{}

	l, brw := NewBrowser(
		BrowserNoDefaultDevice(false),
	)

	bot := NewBot(
		Headless(true),
		UserAgent(userAgent),
		AcceptLanguage(lang),
		Launcher(l),
		Browser(brw),
	)
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL + "/headers")

	elems := bot.page.MustElements(`ul>li`)
	for _, elem := range elems {
		headers = append(headers, elem.MustText())
	}

	s.Contains(headers, fmt.Sprintf("User-Agent: %s", userAgent), "user-agent must match")
	s.Contains(headers, fmt.Sprintf("Accept-Language: %s;q=0.9", lang), "accept-language must match")
}

const (
	_testCkStr = `sessionid=sessionid001; tz=Asia/Shanghai; logged_in=yes`
)

func (s *BotSuite) Test02CopyAsCURLCookies() {
	s.T().Parallel()

	tests := []struct {
		name  string
		ckKey string
		want  string
	}{
		{
			name:  "lowercase key",
			ckKey: "cookie",
			want:  _testCkStr,
		},
		{
			name:  "Canonical key",
			ckKey: "Cookie",
			want:  _testCkStr,
		},
	}
	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			raw := fmt.Sprintf(`curl 'http://127.0.0.1' -H '%s: %s'`, tt.ckKey, _testCkStr)

			bot := NewBotDefault(
				Headless(true),
				CopyAsCURLCookies([]byte(raw)),
			)
			defer bot.Cleanup()

			bot.MustOpen(s.ts.URL + "/check_cookie")
			got := bot.page.MustElement(`pre`).MustText()
			s.Equal(tt.want, got, tt.name)
		})
	}
}

func (s *BotSuite) Test02DumpCookies() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL + "/set_cookie")

	ckFile, err := bot.DumpCookies()
	s.Nil(err)

	defer func() {
		err := os.Remove(ckFile)
		s.Nil(err)
	}()

	nodes, err := bot.LoadCookies(ckFile)
	s.Nil(err)
	s.Len(nodes, 3)

	wanted := strings.Split(_testCkStr, "; ")
	got := flattenNodes(nodes)

	s.ElementsMatch(wanted, got)
}

func (s *BotSuite) Test02WithCookies() {
	s.T().Parallel()

	tests := []struct {
		name string
		bot  *Bot
		want string
	}{
		{
			name: "with cookies",
			bot:  NewBotHeadless(WithCookies(true)),
			want: ".cookies/https_ip.42.pl.cookies",
		},
		{
			name: "with cookie file",
			bot:  NewBotHeadless(WithCookieFile("/tmp/specified_cookie")),
			want: "/tmp/specified_cookie",
		},
		{
			name: "with cookie folder",
			bot:  NewBotHeadless(WithCookieFolder("/tmp/specified")),
			want: "/tmp/specified/https_ip.42.pl.cookies",
		},
	}

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.bot.Cleanup()
			tt.bot.MustOpen(_ip42pl)

			ckFile, err := tt.bot.DumpCookies()
			s.Nil(err)

			defer func() {
				err := os.Remove(ckFile)
				s.Nil(err)
			}()

			s.Equal(tt.want, ckFile)
		})
	}
}

// checkProxyServer
// how to:
// refer: https://docs.mitmproxy.org/stable/overview-getting-started/
// NotUsed: docker run --rm -it -v ~/.mitmproxy:/home/mitmproxy/.mitmproxy -p 8080:8080 mitmproxy/mitmproxy
// docker run --rm -it -p 8080:8080 mitmproxy/mitmproxy mitmdump
// docker run --rm -it -p 8081:8080 mitmproxy/mitmproxy mitmdump --set proxyauth=ab:cd
func checkProxyServer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()
	return nil
}

func (s *BotSuite) Test03ProxyServerByMitm() {
	// s.T().Parallel()
	const (
		// Proxy Server from local docker image.
		// refer: https://docs.mitmproxy.org/stable/overview-getting-started/
		// docker run --rm -it -p 8080:8080 mitmproxy/mitmproxy mitmdump
		// docker run --rm -it -p 8081:8080 mitmproxy/mitmproxy mitmdump --set proxyauth=ab:cd
		// addr = "127.0.0.1:8080"
		httploc  = `h3.my-4`
		httpsloc = `img[alt="mitmproxy"]`
	)

	container := fixtures.NewContainer(context.Background(), "proxy-server", "mitmdump")
	defer func() {
		_ = recover()
		container.Clear()
	}()

	addr := container.URI
	log.Printf("running proxy at: %s", addr)

	containerWithAuth := fixtures.NewContainer(context.Background(), "authed-proxy-server", "mitmdump", "--set", "proxyauth=ab:cd")
	defer func() {
		_ = recover()
		containerWithAuth.Clear()
	}()

	addrAuth := containerWithAuth.URI
	log.Printf("running proxy-auth at: %s", addrAuth)

	// sleep a bit time, in case proxy is not fully loaded.
	SleepN(2.0)
	err := checkProxyServer(addr)
	s.Nil(err, "proxy-server should open")

	err = checkProxyServer(addrAuth)
	s.Nil(err, "authed-proxy-server should open")

	log.Printf("running test")

	tests := []struct {
		name    string
		home    string
		newBrw  func() (*launcher.Launcher, *rod.Browser)
		locator string
		wantE   bool
	}{
		{
			name:    "no proxy",
			home:    _mitmHTTPHome,
			locator: httploc,
			wantE:   true,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				return NewBrowser()
			},
		},
		{
			name:    "with proxy",
			home:    _mitmHTTPHome,
			locator: httploc,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				return NewBrowser(
					BrowserProxy(addr),
				)
			},
		},
		{
			name:    "with proxy for https",
			home:    _mitmHTTPSHome,
			locator: httpsloc,
			wantE:   true,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				return NewBrowser(
					BrowserProxy(addr),
				)
			},
		},
		{
			name:    "with proxy for https ignore cert err",
			home:    _mitmHTTPSHome,
			locator: httpsloc,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				return NewBrowser(
					BrowserProxy(addr),
					BrowserIgnoreCertErrors(true),
				)
			},
		},
		{
			name:    "with proxy extension",
			home:    _mitmHTTPHome,
			locator: httploc,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				extPth, err := NewChromeExtension(addrAuth+":ab:cd:", "/tmp")
				dry.PanicIfErr(err)

				return NewBrowser(BrowserExtensions(extPth, _testExtension))
			},
		},
		{
			name:    "with non-proxy extension",
			home:    _mitmHTTPHome,
			locator: httploc,
			wantE:   true,
			newBrw: func() (*launcher.Launcher, *rod.Browser) {
				return NewBrowser(BrowserExtensions(_testExtension))
			},
		},
	}

	swg := sizedwaitgroup.New(3)

	for _, tt := range tests {
		swg.Add()
		go func() {
			defer swg.Done()

			l, brw := tt.newBrw()
			bot := NewBot(Launcher(l), Browser(brw))
			defer bot.Cleanup()

			_ = bot.Open(tt.home)
			_, err := bot.Elem(tt.locator, WithTimeout(3))

			if tt.wantE {
				s.NotNil(err, tt.name)
			} else {
				s.Nil(err, tt.name)
			}
		}()
	}

	swg.Wait()
}

func (s *BotSuite) ProxyWithAuthExample() {
	proxy := "ip:port"

	l, brw := NewBrowser(BrowserProxy(proxy))
	bot := NewBot(Launcher(l), Browser(brw))
	go brw.MustHandleAuth("username", "password")()

	bot.MustOpen(_ip42pl)

	ipaddr := bot.MustElemAttr(`p.ip`)
	fmt.Printf("got ip: %s\n", ipaddr)

	QuitOnTimeout(2)
}

func (s *BotSuite) Test04Eval() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL)

	got1 := bot.MustEval(`() => console.log("hello world")`)
	s.Equal("<nil>", got1)

	got2 := bot.Page().MustEval(`(a, b) => a + b`, 1, 2).Int()
	s.Equal(3, got2)

	_, err := bot.Eval(`() => throw "too big"`)
	var evalErr *rod.EvalError
	s.ErrorAs(err, &evalErr, "should throw error")
}

func (s *BotSuite) Test04URLContains() {
	s.T().Parallel()

	tests := []struct {
		name     string
		expected string
		wantE    bool
	}{
		{
			name:     "found delay",
			expected: `activate/delay`,
		},
		{
			name:     "should not found detail 2",
			expected: `activate/detail`,
			wantE:    true,
		},
	}

	home := s.ts.URL + "/activate"

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			bot := NewBotHeadless(TrackTime(true), Logger(zlog.MustNewLoggerDebug()))
			defer bot.Cleanup()

			bot.MustOpen(home)
			bot.MustClick(`a[href="/activate/delay"]`)

			err := bot.WaitURLContains(tt.expected, 3.0)
			if tt.wantE {
				s.NotNil(err)
			} else {
				s.Nil(err)
				s.Equal(home+"/delay", bot.CurrentURL())
			}
		})
	}
}

func (s *BotSuite) Test04ActivateLastOpenedPage() {
	s.T().Parallel()

	const retry = 3

	tests := []struct {
		name       string
		reg        string
		activateOp func(*Bot, rod.Pages, string) error
		wantE      bool
	}{
		{
			name: "by last opened page",
			activateOp: func(bot *Bot, pages rod.Pages, _ string) error {
				return bot.ActivateLastOpenedPage(pages, retry)
			},
		},
		{
			name: "by regex",
			reg:  "newtab",
			activateOp: func(bot *Bot, _ rod.Pages, reg string) error {
				return bot.ActivatePageByURLRegex(reg, retry)
			},
		},
		{
			name: "by regex",
			reg:  "current",
			activateOp: func(bot *Bot, _ rod.Pages, reg string) error {
				return bot.ActivatePageByURLRegex(reg, retry)
			},
			wantE: true,
		},
	}
	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			bot := NewBotHeadless(TrackTime(true))
			defer bot.Cleanup()

			bot.MustOpen(s.ts.URL + "/activate")

			s.Equal("Activate", bot.page.MustInfo().Title, tt.name)

			pages := bot.browser.MustPages()

			bot.MustClick(`a[href="/activate/newtab"]`)
			// when open new tab click, bot focused page is not changed by default.
			s.Equal("Activate", bot.page.MustInfo().Title, tt.name)

			// activate the last opened page.
			// bot.ActivateLastOpenedPage(pages, 10)
			err := tt.activateOp(bot, pages, tt.reg)
			if tt.wantE {
				s.NotNil(err, tt.name)
			} else {
				s.Nil(err, tt.name)
				s.Equal("HelloNewTab", bot.page.MustInfo().Title, tt.name)

				err = bot.ResetToOriginalPage()
				s.Nil(err)
			}
		})
	}
}

func (s *BotSuite) TestMustDOMStable() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL + "/dynamic")

	start := time.Now()
	bot.MustDOMStable()
	elapsed := time.Since(start)

	s.GreaterOrEqual(elapsed, time.Second)
	s.True(bot.page.MustHas("div#dynamic-content"))
}

func (s *BotSuite) TestMustWaitLoad() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	start := time.Now()
	bot.MustOpen(s.ts.URL + "/slow-load")
	bot.MustWaitLoad()
	elapsed := time.Since(start)

	s.GreaterOrEqual(elapsed, 2*time.Second)
	s.True(bot.page.MustHas("div#loaded-content"))
}

func (s *BotSuite) TestEnsureCookieFile() {
	s.T().Parallel()

	bot := NewBotHeadless(WithCookies(true))
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL)
	err := bot.ensureCookieFile()
	s.Nil(err)
	s.NotEmpty(bot.cookieFile)
	_, err = bot.DumpCookies()
	s.Nil(err)
	s.FileExists(bot.cookieFile)

	// Clean up
	os.Remove(bot.cookieFile)
}

func (s *BotSuite) TestSetPanicWith() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	bot.SetPanicWith(PanicByDump)
	s.Equal(PanicByDump, bot.panicBy)

	bot.SetPanicWith(PanicByLogError)
	s.Equal(PanicByLogError, bot.panicBy)
}

func (s *BotSuite) TestTimeTrack() {
	s.T().Parallel()

	bot := NewBotHeadless(TrackTime(true))
	defer bot.Cleanup()

	// Capture log output
	var buf bytes.Buffer
	bot.logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&buf),
		zap.DebugLevel,
	))

	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	bot.TimeTrack(start, 1)

	s.Contains(buf.String(), "took")
	s.Contains(buf.String(), "TestTimeTrack")
}

func (s *BotSuite) TestPieErrorHandling() {
	s.T().Parallel()

	testError := errors.New("test error") //nolint

	tests := []struct {
		name      string
		panicBy   PanicByType
		wantPanic bool
	}{
		{"PanicByDump", PanicByDump, false},
		{"PanicByLogError", PanicByLogError, false},
		{"PanicByDft", PanicByDft, true},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			bot := NewBotHeadless()
			defer bot.Cleanup()

			bot.SetPanicWith(tt.panicBy)

			if tt.wantPanic {
				s.Panics(func() { bot.pie(testError) })
			} else {
				s.NotPanics(func() { bot.pie(testError) })
			}
		})
	}
}

func (s *BotSuite) TestNewBotWithOptionsOnly() {
	s.T().Parallel()

	bot := NewBotWithOptionsOnly(WithHighlightTimes(5), UserAgent("TestAgent"))
	s.Nil(bot.page)
	s.Equal(5, bot.highlightTimes)
	s.Equal("TestAgent", bot.userAgent)
}

func (s *BotSuite) TestBlockInCleanUp() {
	s.T().Parallel()

	bot := NewBotHeadless()

	done := make(chan bool)
	go func() {
		bot.BlockInCleanUp()
		done <- true
	}()

	select {
	case <-done:
		s.Fail("BlockInCleanUp should not return")
	case <-time.After(2 * time.Second):
		// This is expected behavior
	}
}

func (s *BotSuite) TestSetTimeout() {
	s.T().Parallel()

	bot := NewBotHeadless()
	bot.SetTimeout()

	s.Equal(LongToSec*time.Second, bot.longTimeout)
	s.Equal(MediumToSec*time.Second, bot.mediumTimeout)
	s.Equal(ShortToSec*time.Second, bot.shortTimeout)
	s.Equal(NapToSec*time.Second, bot.napTimeout)
	s.Equal(PT10Sec*time.Second, bot.pt10s)
	s.Equal(1*time.Second, bot.pt1s)
}

func (s *BotSuite) TestCurrentURL() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	bot.MustOpen(s.ts.URL)
	s.Equal(s.ts.URL, strings.TrimSuffix(bot.CurrentURL(), "/"))
}

func (s *BotSuite) TestResetIfHeadless() {
	s.T().Parallel()

	bot := NewBot(Headless(true))
	originalBrowser := bot.browser

	resetIfHeadless(bot)

	s.NotEqual(originalBrowser, bot.browser, "Browser should be reset when headless is true")
}

func (s *BotSuite) TestNewBotForDebug() {
	s.T().Parallel()

	bot := NewBotForDebug()
	defer bot.Cleanup()

	s.NotNil(bot.launcher)
	s.NotNil(bot.browser)
	// Check if the --show-paint-rects flag is set
	pr := PaintRects[2:]

	found := false
	for k := range bot.launcher.Flags {
		if pr == string(k) {
			found = true
			break
		}
	}

	s.True(found)
}

func (s *BotSuite) TestInitialize() {
	s.T().Parallel()

	bot := &Bot{}
	bot.initialize()

	s.NotNil(bot.logger)
	s.Equal(1, bot.highlightTimes)
	s.Len(bot.UniqueID, _uniqueIDLen)
}

func (s *BotSuite) TestBindBotLauncher() {
	s.T().Parallel()

	bot := NewBotWithOptionsOnly()
	s.False(bot.isLaunched)
	s.Nil(bot.page)

	BindBotLanucher(bot)
	s.True(bot.isLaunched)
	s.NotNil(bot.page)

	// Test that calling BindBotLanucher again doesn't change anything
	originalPage := bot.page
	BindBotLanucher(bot)
	s.Equal(originalPage, bot.page)
}

func (s *BotSuite) TestOverrideUA() {
	s.T().Parallel()

	ua := "TestUserAgent"
	lang := "en-US"
	override := overrideUA(ua, lang)

	s.Equal(ua, override.UserAgent)
	s.Equal(lang, override.AcceptLanguage)

	// Test default language
	override = overrideUA(ua, "")
	s.Equal("en-CN,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,en-GB;q=0.6,en-US;q=0.5", override.AcceptLanguage)
}

func (s *BotSuite) TestBotOptions() {
	s.T().Parallel()

	testUA := "TestUserAgent"
	testLang := "zh-CN"
	testDir := "/tmp/test_user_data"
	testCookieFile := "/tmp/test_cookies.json"

	bot := NewBot(
		UserAgent(testUA),
		AcceptLanguage(testLang),
		UserDataDir(testDir),
		WithHighlightTimes(5),
		WithCookieFile(testCookieFile),
		Humanized(true),
		StealthMode(true),
		ForceCleanup(true),
		ClearCookies(true),
		TrackTime(true),
		WithLeftPosition(100),
	)
	defer bot.Cleanup()

	s.Equal(testUA, bot.userAgent)
	s.Equal(testLang, bot.acceptLanguage)
	s.Equal(testDir, bot.userDataDir)
	s.Equal(5, bot.highlightTimes)
	s.Equal(testCookieFile, bot.cookieFile)
	s.True(bot.humanized)
	s.True(bot.stealthMode)
	s.True(bot.forceCleanup)
	s.True(bot.clearCookies)
	s.True(bot.trackTime)
	s.Equal(100, bot.left)
}

func (s *BotSuite) TestLogTimeSpent() {
	s.T().Parallel()

	var buf bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&buf),
		zap.DebugLevel,
	))

	bot := NewBotHeadless(Logger(logger), TrackTime(true))
	defer bot.Cleanup()

	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	bot.LogTimeSpent(start)

	s.Contains(buf.String(), "TestLogTimeSpent")
	s.Contains(buf.String(), "took")
}

func (s *BotSuite) TestClearStorageCookies() {
	s.T().Parallel()

	bot := NewBotHeadless()
	defer bot.Cleanup()

	// Set a cookie
	bot.MustOpen(s.ts.URL + "/set_cookie")

	// Clear cookies
	bot.ClearStorageCookies()

	// Verify cookies are cleared
	cookies, err := bot.browser.GetCookies()
	s.Nil(err)
	s.Empty(cookies)
}
