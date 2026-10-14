package main

import (
	"fmt"

	"github.com/coghost/pathlib"
	"github.com/coghost/sleep"
	"github.com/coghost/wee"
	rod "github.com/headlesslab/wand"
	"github.com/headlesslab/wand/lib/launcher"
	"github.com/k0kubun/pp/v3"
)

func main() {
	lch := launcher.NewUserMode().
		UserDataDir("/tmp/cb1002/1021").
		Set("disable-default-apps").
		Set("no-first-run")

	lch = lch.Bin("/Applications/Google Chrome for Testing.app/Contents/MacOS/Google Chrome For Testing")
	lch = lch.Headless(false)
	lch = lch.Leakless(true)
	lch.Set("load-extension", pathlib.Path("~/tmp/extensions/site805/superproxyv3").ExpandUser().String())

	wsURL := lch.MustLaunch()

	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	options := []wee.BotOption{
		wee.Headless(false),
		wee.Devtools(true),
		wee.WithBounds(wee.NewBoundsHD()),
		wee.Launcher(lch),
		wee.Browser(browser),
	}

	bot := wee.NewBot(options...)
	defer wee.Blocked()

	sleep.PT3s()
	// bot.MustOpen("https://ifconfig.co/json")
	wee.CurlIPFromIfconfigCO(bot)
	sleep.PT3s()

	bot.MustOpen("https://www.crunchbase.com/organization/indeed/signals_and_news/timeline")

	elem, err := bot.Elem(`input[name="cf-turnstile-response"]`)
	if err != nil {
		fmt.Printf("got error: %v\n", err)
	}

	if elem == nil {
		pp.Println("no iframe found")
		return
	}
}
