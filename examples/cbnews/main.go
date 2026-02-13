package main

import (
	"fmt"
	"log"

	"github.com/coghost/pathlib"
	"github.com/coghost/sleep"
	"github.com/coghost/wee"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const home = "https://www.crunchbase.com"

var newsLayout = "https://www.crunchbase.com/organization/%s/news_and_analysis"

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

	bot.MustOpen(home)
	sleep.PT3s()
	bot.MustOpen(fmt.Sprintf(newsLayout, "indeed"))
	sleep.PT3s()

	elem, err := bot.Elem(`button[aria-label="Load More"]`, wee.WithTimeout(5.0))
	if err != nil {
		log.Printf("find elem has error: %v", err)
		return
	}

	bot.ScrollToElement(elem)
	sleep.PT1s()

	if err := bot.ClickElem(elem, wee.WithClickByScript(true)); err != nil {
		log.Printf("cannot click elem: %v", err)
		return
	}
}
