package wee

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/samber/lo"
)

const PaintRects = "--show-paint-rects"

func NewBrowser(opts ...BrowserOptionFunc) (*launcher.Launcher, *rod.Browser) {
	delay := 500
	opt := BrowserOptions{slowDelay: delay}
	bindBrowserOptions(&opt, opts...)

	l := NewLauncher(opts...)
	brw := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	brw.SlowMotion(time.Millisecond * time.Duration(opt.slowDelay))

	return l, brw
}

func NewLauncher(opts ...BrowserOptionFunc) *launcher.Launcher {
	opt := BrowserOptions{}
	bindBrowserOptions(&opt, opts...)

	lauch := launcher.New()
	setLauncher(lauch, false)

	if opt.paintRects {
		lauch.Set("--show-paint-rects")
	}

	if len(opt.flags) != 0 {
		lo.ForEach(opt.flags, func(item string, i int) {
			lauch.Set(flags.Flag(item))
		})
	}

	return lauch
}

func setLauncher(client *launcher.Launcher, headless bool) {
	client.
		Set("no-first-run").
		Set("no-startup-window").
		Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("disable-web-security").
		Delete("use-mock-keychain").
		Set("disable-infobars").
		Set("enable-automation").
		NoSandbox(true).
		Headless(headless)
}

type BrowserOptions struct {
	slowDelay  int
	paintRects bool
	flags      []string
}

type BrowserOptionFunc func(o *BrowserOptions)

func bindBrowserOptions(opt *BrowserOptions, opts ...BrowserOptionFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithPaintRects(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.paintRects = b
	}
}

func WithSlowDelay(i int) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.slowDelay = i
	}
}

func WithFlags(arr ...string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.flags = append(o.flags, arr...)
		o.flags = lo.Uniq(o.flags)
	}
}
