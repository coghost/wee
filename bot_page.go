package wee

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

const (
	ActivatePageRetry = 10
)

const (
	_domStableDiff = 0.95
)

var (
	ErrCannotActivateOpenedPage = errors.New("cannot activate latest opened page")
	ErrMissingCookieFile        = errors.New("missing cookie file")
)

func (b *Bot) CustomizePage() {
	if !b.withPageCreation {
		return
	}

	if b.launcher == nil {
		return
	}

	if b.page == nil {
		if b.stealthMode && !b.userMode {
			b.page = stealth.MustPage(b.browser)
		} else {
			b.page = b.browser.MustPage()
		}
	}

	ua := b.userAgent
	lang := b.acceptLanguage

	if ua != "" || lang != "" {
		ov := overrideUA(ua, lang)
		b.page = b.page.MustSetUserAgent(ov)
	}

	// display := NewDefaultDisplay()
	// if b.display != nil {
	// 	display = b.display
	// }
	// w, h := 1280, 768
	// b.page = b.page.MustSetWindow(display.Left, display.Top, display.Width, display.Height)
	// vw, vh := w, 728
	// b.page = b.page.MustSetViewport(display.ViewOffsetWidth, display.ViewOffsetHeight, 0.0, false)

	err := b.page.SetWindow(&proto.BrowserBounds{
		Left:        &b.left,
		WindowState: proto.BrowserWindowStateNormal,
	})
	if err != nil {
		panic(err)
	}

	_ = b.page.SetViewport(nil)

	b.isLaunched = true
}

func (b *Bot) MustOpen(uri string) {
	defer b.LogTimeSpent(time.Now())

	withCookies := b.withCookies ||
		b.cookieFile != "" ||
		len(b.copyAsCURLCookies) != 0

	b.logger.Info("open page", zap.Bool("cookies", withCookies))
	// not with cookies, return
	if !withCookies {
		b.pie(b.Open(uri))
		return
	}

	b.pie(b.OpenAndSetCookies(uri))
	b.pie(b.Open(uri))
}

// OpenAndSetCookies open uri with cookies.
//
// typically with following steps:
//   - open it's domain `https://xxx.com`
//   - load cookies
//   - open uri
func (b *Bot) OpenAndSetCookies(uri string, timeouts ...time.Duration) error {
	up, err := url.Parse(uri)
	if err != nil {
		return err
	}

	homepage := fmt.Sprintf("%s://%s", up.Scheme, up.Host)
	if err := b.Open(homepage); err != nil {
		return err
	}

	if b.cookieFile == "" {
		_ = b.ensureCookieFile()
		b.logger.Sugar().Infof("use cookie file %s", b.cookieFile)
	}

	nodes, err := b.LoadCookies(b.cookieFile)
	if err != nil {
		b.logger.Info("cannot load cookies", zap.String("cookie", b.cookieFile), zap.Error(err))
	}

	if len(nodes) != 0 {
		b.logger.Sugar().Infof("total got cookie-nodes: %d", len(nodes))
		// set cookies to browser/page
		// if err = b.page.SetCookies(nodes); err != nil {
		// 	b.logger.Error("set cookies failed", zap.Error(err))
		// }
		if err = b.browser.SetCookies(nodes); err != nil {
			b.logger.Error("set cookies failed", zap.Error(err))
		}
	}

	return nil
}

func (b *Bot) Open(url string, timeouts ...time.Duration) error {
	timeout := FirstOrDefault(b.longTimeout, timeouts...)

	if err := b.page.Timeout(timeout).Navigate(url); err != nil {
		return err
	}

	return b.page.Timeout(timeout).WaitLoad()
}

// OpenURLInNewTab opens url in a new tab and activates it
func (b *Bot) OpenURLInNewTab(uri string) error {
	p, err := b.browser.Page(proto.TargetCreateTarget{URL: uri})
	if err != nil {
		return fmt.Errorf("cannot new page with url(%s): %w", uri, err)
	}

	if err := b.ActivatePage(p); err != nil {
		return fmt.Errorf("cannot activate new page: %w", err)
	}

	return nil
}

// MustEval  a wrapper with MediumTo to rod.Page.MustEval
//
// refer: https://github.com/go-rod/rod/blob/main/examples_test.go#L53
//
//	@param script `() => console.log("hello world")` or "`(a, b) => a + b`, 1, 2"
//	@return string
func (b *Bot) MustEval(script string) string {
	res := b.page.Timeout(b.mediumTimeout).MustEval(script).String()
	return res
}

func (b *Bot) Eval(script string) (*proto.RuntimeRemoteObject, error) {
	obj, err := b.page.Timeout(b.mediumTimeout).Eval(script)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// WaitURLContains uses `decodeURIComponent(window.location.href).includes` to check if url has str or not,
// Scenario:
//
//   - if a page's loading status can be determined by url
func (b *Bot) WaitURLContains(str string, timeouts ...float64) error {
	defer b.LogTimeSpent(time.Now())

	timeout := FirstOrDefault(float64(MediumToSec), timeouts...)
	script := fmt.Sprintf(`() => decodeURIComponent(window.location.href).includes("%s")`, str)

	return rod.Try(func() {
		b.page.Timeout(time.Second * time.Duration(timeout)).MustWait(script).CancelTimeout()
	})
}

func (b *Bot) MustDOMStable() {
	defer b.LogTimeSpent(time.Now())

	b.pie(b.DOMStable(b.pt1s, _domStableDiff))
}

func (b *Bot) DOMStable(d time.Duration, diff float64) error {
	return b.page.Timeout(b.mediumTimeout).WaitDOMStable(d, diff)
}

func (b *Bot) MustWaitLoad() {
	b.page.Timeout(b.mediumTimeout).MustWaitLoad().CancelTimeout()
}

// ActivatePage activates a page instead of current.
func (b *Bot) ActivatePage(page *rod.Page) error {
	b.prevPage, b.page = b.page, page
	_, err := b.page.Activate()

	return err
}

func (b *Bot) MustActivatePage(page *rod.Page) {
	b.pie(b.ActivatePage(page))
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
	defer b.LogTimeSpent(time.Now())

	var page *rod.Page

	for i := 0; i < retry; i++ {
		page, _ = b.browser.MustPages().FindByURL(jsRegex)
		if page == nil {
			SleepN(1.0)
			continue
		}

		break
	}

	if page == nil {
		return ErrCannotActivateOpenedPage
	}

	return b.ActivatePage(page)
}

func (b *Bot) ActivateLastOpenedPage(pagesBefore rod.Pages, retry int) error {
	defer b.LogTimeSpent(time.Now())

	var page *rod.Page

	for i := 0; i < retry; i++ {
		pagesAfter := b.browser.MustPages()
		for _, p := range pagesAfter {
			if !funk.Contains(pagesBefore, p) {
				page = p
				break
			}
		}

		if page != nil {
			break
		}

		SleepN(1.0)
	}

	if page == nil {
		return ErrCannotActivateOpenedPage
	}

	return b.ActivatePage(page)
}

func (b *Bot) LogTimeSpent(start time.Time, skips ...int) {
	const (
		defaultSkip = 2
	)

	skip := FirstOrDefault(defaultSkip, skips...)

	if b.trackTime {
		b.TimeTrack(start, skip)
	}
}

func (b *Bot) TimeTrack(start time.Time, skip int) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, file, _, _ := runtime.Caller(skip)
	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)
	fname := filepath.Base(file)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	b.logger.Sugar().Debugf("%s.%s took %.2fs", fname, name, elapsed.Seconds())
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
