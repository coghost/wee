package main

import (
	"context"
	"errors"
	"fmt"

	"wee"
	"wee/schemer"

	"github.com/coghost/xlog"
	"github.com/coghost/xpretty"
	"github.com/k0kubun/pp/v3"
)

var blocket = schemer.Scheme{
	Uid: "blocket.se",
	Mapper: &schemer.Mapper{
		Home:         "https://jobb.blocket.se/",
		Popovers:     []string{`iframe[id^="sp_message_iframe"]$$$button[class~=sp_choice_type_11]`},
		SortBy:       schemer.NewDropdown(`div#sort-filter`, `div[data-value="sp=%s"]`),
		HasResults:   []string{`div.job-item a.header`},
		ResultsCount: `span[id="total-count"]`,
		NextPage:     `div.pagination a>i.right`,
		PageNum:      `button.blue`,
	},
}

func main() {
	xlog.InitLogDebug()
	xpretty.InitializeWithColor()

	selectors := blocket.Mapper

	bot := wee.NewBotDefault(
		wee.WithPopovers(selectors.Popovers...),
		wee.WithPanicBy(wee.PanicByDump),
		wee.WithHighlightTimes(1),
	)

	defer bot.Cleanup()
	defer wee.Blocked()

	bot.MustOpen(blocket.Mapper.Home)

	bot.MustAcceptCookies(selectors.Popovers...)

	// bot.MustClickElem(`div.all-categories span.more`, wee.WithClickByScript(true))
	// bot.MustClickAndWait(`div.title a[href$="juridik/"] label`, wee.WithClickByScript(true))
	bot.ClickWithScript(`div.title a[href$="juridik/"] label`)

	val := bot.MustElemAttr(selectors.ResultsCount)
	pp.Println(val)

	bot.MustClickOneByOne(selectors.SortBy.Trigger, fmt.Sprintf(selectors.SortBy.Item, "0"))

	for i := 0; i < 0; i++ {
		pageNum := bot.MustElemAttr(selectors.PageNum)
		pp.Println("current page", pageNum)

		results := bot.MustElemsForAllSelectors(selectors.HasResults)
		pp.Println("total jobs", len(results))

		titles := bot.MustAllElemsAttrs(selectors.HasResults[0], wee.WithAttr("href"))
		pp.Println(titles)

		elems := bot.MustElemsForAllSelectors(selectors.HasResults)
		res := bot.AllElementsAttrMap(elems, wee.WithAttrMap(map[string]string{
			"url":   "href",
			"title": "",
			"class": "class",
		}))
		pp.Println(res)

		button, err := bot.Elem(selectors.NextPage, wee.WithTimeout(wee.NapToSec))
		if errors.Is(err, context.DeadlineExceeded) {
			pp.Println("no more next page button found")
			bot.MustScrollToBottom()

			break
		}

		bot.MustClickElemAndWait(button)
	}

	url := bot.CurrentUrl()
	pp.Println(url)
}
