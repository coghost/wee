package wee

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/gookit/goutil/fsutil"
	"github.com/k0kubun/pp/v3"
	"github.com/samber/lo"
)

const (
	// SlowMotionMillis is the default slow motion duration in milliseconds
	SlowMotionMillis = 500
)

// PaintRects is a flag to show paint rectangles in the browser
const PaintRects = "--show-paint-rects"

func NewUserMode(opts ...BrowserOptionFunc) (*launcher.Launcher, *rod.Browser) {
	opt := BrowserOptions{}
	bindBrowserOptions(&opt, opts...)

	lch, wsURL := newUserModeLauncher(opts...)
	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	log.Printf("running in user-mode with user-data-dir:(%s)", opt.userDataDir)

	return lch, browser
}

func NewBrowser(opts ...BrowserOptionFunc) (*launcher.Launcher, *rod.Browser) {
	opt := BrowserOptions{slowMotionDelay: SlowMotionMillis, noDefaultDevice: true}
	bindBrowserOptions(&opt, opts...)

	lnchr := NewLauncher(opts...)

	brw := rod.New().ControlURL(lnchr.MustLaunch()).MustConnect()
	if opt.noDefaultDevice {
		brw.NoDefaultDevice()
	}

	if opt.incognito {
		brw.MustIncognito()
	}

	// just ignore cert errors
	if opt.ignoreCertErrors {
		_ = brw.IgnoreCertErrors(opt.ignoreCertErrors)
	}

	brw.SlowMotion(time.Millisecond * time.Duration(opt.slowMotionDelay))

	return lnchr, brw
}

func NewLauncher(opts ...BrowserOptionFunc) *launcher.Launcher {
	opt := BrowserOptions{}
	bindBrowserOptions(&opt, opts...)

	lnchr := launcher.New()
	setLauncher(lnchr, opt.headless)

	// for _, extFolder := range opt.extensions {
	lnchr.Set("load-extension", strings.Join(opt.extensions, ","))
	// }

	if dir := opt.userDataDir; dir != "" {
		lnchr.UserDataDir(fsutil.Expand(dir))
	}

	if opt.paintRects {
		lnchr.Set(PaintRects)
	}

	if len(opt.flags) != 0 {
		lo.ForEach(opt.flags, func(item string, i int) {
			lnchr.Set(flags.Flag(item))
		})
	}

	if proxy := opt.proxy; proxy != "" {
		lnchr.Proxy(proxy)
	}

	return lnchr
}

func newUserModeLauncher(opts ...BrowserOptionFunc) (*launcher.Launcher, string) {
	opt := BrowserOptions{}
	bindBrowserOptions(&opt, opts...)

	launch := launcher.NewUserMode()

	if dir := opt.userDataDir; dir != "" {
		launch = launch.UserDataDir(fsutil.Expand(dir))
		pp.Println("set user-data-dir to:", dir)
	}

	launch.Leakless(opt.leakless)

	wsURL, err := launch.Launch()
	if err != nil {
		s := fmt.Sprintf("%s", err)
		if strings.Contains(s, "[launcher] Failed to get the debug url: Opening in existing browser session") {
			fmt.Printf("%[1]s\nlaunch chrome browser failed, please make sure chrome is closed, and then run again\n%[1]s\n", strings.Repeat("=", 32)) //nolint
		}

		log.Fatalf("cannot launch browser: %v", err)
	}

	return launch, wsURL
}

func setLauncher(client *launcher.Launcher, headless bool) {
	// Delete("use-mock-keychain"). if add this, a popup with message: "chromium wants to use your confidential information" will shown, and you have to manually confirm it.
	client.
		Set("no-sandbox").
		Set("no-first-run").
		Set("no-startup-window").
		Set("disable-blink-features", "AutomationControlled").
		// Set("disable-gpu").
		Set("disable-dev-shm-usage").
		// Set("disable-web-security").
		Set("disable-infobars").
		Set("enable-automation").
		Headless(headless)
}

// func loadProxyExtension(l *launcher.Launcher, proxyLine string) {
// 	extensionFolder, _ := NewChromeExtension(proxyLine, "/tmp")
// 	l.Set("load-extension", extensionFolder)
// 	log.Info().Str("extension_folder", extensionFolder).Msg("load proxy extension")
// }
