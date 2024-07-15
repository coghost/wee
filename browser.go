package wee

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	// SlowMotionMillis default slow motion duration 500ms
	SlowMotionMillis = 500
)

const PaintRects = "--show-paint-rects"

func NewUserMode() (*launcher.Launcher, *rod.Browser) {
	l, wsURL := newUserModeLauncher()
	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	return l, browser
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

	brw.SlowMotion(time.Millisecond * time.Duration(opt.slowMotionDelay))

	return lnchr, brw
}

func NewLauncher(opts ...BrowserOptionFunc) *launcher.Launcher {
	opt := BrowserOptions{}
	bindBrowserOptions(&opt, opts...)

	lnchr := launcher.New()
	setLauncher(lnchr, opt.headless)

	for _, extFolder := range opt.extensions {
		lnchr.Set("load-extension", extFolder)
	}

	if dir := opt.userDataDir; dir != "" {
		lnchr.UserDataDir(dir)
	}

	if opt.paintRects {
		lnchr.Set(PaintRects)
	}

	if len(opt.flags) != 0 {
		lo.ForEach(opt.flags, func(item string, i int) {
			lnchr.Set(flags.Flag(item))
		})
	}

	return lnchr
}

func newUserModeLauncher() (*launcher.Launcher, string) {
	launch := launcher.NewUserMode()

	wsURL, err := launch.Launch()
	if err != nil {
		s := fmt.Sprintf("%s", err)
		if strings.Contains(s, "[launcher] Failed to get the debug url: Opening in existing browser session") {
			fmt.Printf("%[1]s\nlaunch chrome browser failed, please make sure it is closed, and then run again\n%[1]s\n", strings.Repeat("=", 32)) //nolint
		}

		log.Fatal().Err(err).Msg("cannot launch browser")
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
