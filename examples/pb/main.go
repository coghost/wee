package main

import (
	"fmt"
	"log"

	"github.com/coghost/pathlib"
	"github.com/coghost/sleep"
	"github.com/coghost/wee"
	"github.com/coghost/xdtm"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	lch := launcher.NewUserMode().
		UserDataDir(pathlib.Path("/tmp/pb").ExpandUser().String()).
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

	// bot.MustOpen("https://www.zhixinglib.com/lib/")
	if !wee.Confirm("Please open url and devtools, then process enter") {
		return
	}

	nextPageButtonLocator := `button[aria-label="Go to next page"]`
	listingDateLocator := `div[data-id="lastUpdateDate"] span[title]`

	endDate := "20251020"

	for i := range 100 {
		pageNum := i + 1
		nextPageNum := i + 2

		log.Printf("crawling page: %d", pageNum)

		elem := bot.MustElem(nextPageButtonLocator, wee.WithTimeout(wee.PT20Sec))
		bot.MustClickElem(elem)

		dateStr := bot.MustElemAttr(listingDateLocator, wee.WithTimeout(wee.PT10Sec))

		cbobj, err := xdtm.ToCarbon(dateStr, xdtm.WithBySearch(true))
		if err != nil {
			log.Printf("cannot parse date str: %v", err)
			if !wee.Confirm("cannot parse date, wanna continue?") {
				break
			}
		}

		shortDateStr := cbobj.ToShortDateString()
		log.Printf("loaded companies with starting date: %s", shortDateStr)

		sleep.PTXs(60)

		if shortDateStr > endDate {
			continue
		}

		hint := fmt.Sprintf("end date matched wannt go to page %d", nextPageNum)
		if !wee.Confirm(hint) {
			log.Printf("end crawling!")
			break
		}
	}
}
