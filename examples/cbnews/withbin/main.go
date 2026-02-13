package main

import (
	"github.com/coghost/pathlib"
	"github.com/coghost/sleep"
	"github.com/coghost/wee"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	lch := launcher.NewUserMode().
		UserDataDir(pathlib.Path("~/tmp/sb/1021").ExpandUser().String()).
		Set("disable-default-apps").
		Set("no-first-run")

	lch = lch.Bin("/Applications/Google Chrome for Testing.app/Contents/MacOS/Google Chrome For Testing")
	lch = lch.Headless(false)

	wsURL := lch.MustLaunch()

	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	options := []wee.BotOption{
		wee.Headless(false),
		wee.Devtools(true),
		// wee.WithBounds(wee.NewBoundsHD()),
		wee.Launcher(lch),
		wee.Browser(browser),
	}

	bot := wee.NewBot(options...)
	defer wee.Blocked()

	sleep.PT3s()

	bot.MustOpen("https://www.crunchbase.com/")
}
