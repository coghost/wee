package main

import (
	"context"
	"errors"

	"wee"
	"wee/spider"

	"github.com/coghost/xlog"
	"github.com/coghost/xpretty"
	"github.com/k0kubun/pp/v3"
)

var blocket = spider.Spider{
	Uid:  "blocket.se",
	Home: "https://jobb.blocket.se/",
	Selectors: &spider.Selectors{
		Popovers:     []string{`iframe[id^="sp_message_iframe"]$$$button[class~=sp_choice_type_11]`},
		SortBy:       spider.NewDropdown(`div#sort-filter`, `div[data-value="sp=%s"]`),
		HasResults:   []string{`div.job-item a.header`},
		ResultsCount: `span[id="total-count"]`,
		Paginate:     `div.pagination a>i.right`,
		PageNum:      `button.blue`,
	},
}

func main() {
	xlog.InitLogDebug()
	xpretty.InitializeWithColor()

	bot := wee.NewBotWithDefault(
		// wee.WithPopovers(blocket.Webpage.Popovers...),
		wee.WithPanicBy(wee.PanicByDump),
		wee.WithHighlightTimes(0),
	)

	defer bot.Cleanup()
	defer bot.QuitOnTimeout(-1)

	bot.MustOpen(blocket.Home)
	selectors := blocket.Selectors

	bot.MustAcceptCookies(selectors.Popovers...)

	// bot.MustClickElem(selectors.SortBy.Trigger)
	// bot.MustClickElemAndWait(fmt.Sprintf(selectors.SortBy.Item, "0"))

	bot.MustClickElem(`div.all-categories span.more`, wee.WithClickByScript(true))
	wee.RandSleep(0.5, 0.6)
	bot.MustClickElemAndWait(`div.title a[href$="juridik/"] label`)

	val := bot.MustGetElemAttr(selectors.ResultsCount)
	pp.Println(val)

	for i := 0; i < 20; i++ {
		pageNum := bot.MustGetElemAttr(selectors.PageNum)
		pp.Println("current page", pageNum)

		results := bot.MustGetElemsOfAllSelectors(selectors.HasResults)
		pp.Println("total jobs", len(results))

		titles := bot.MustGetAllElemsAttrs(selectors.HasResults[0], wee.WithAttr("href"))
		pp.Println(titles)

		elems := bot.MustGetElemsOfAllSelectors(selectors.HasResults)
		res := bot.GetAllElemsAttrMap(elems, wee.WithAttrMap(map[string]string{
			"url":   "href",
			"title": "",
			"class": "class",
		}))
		pp.Println(res)

		button, err := bot.GetElem(selectors.Paginate, wee.WithTimeout(wee.NapToSec))
		if errors.Is(err, context.DeadlineExceeded) {
			pp.Println("no more next page button found")
			bot.ScrollToBottom()

			break
		}

		bot.MustClickElementAndWait(button)
	}

	url := bot.CurrentUrl()
	pp.Println(url)
}
