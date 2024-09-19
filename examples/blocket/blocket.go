package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/coghost/wee"
	"github.com/coghost/xpretty"
	"github.com/coghost/zlog"
	"go.uber.org/zap"
)

type Settings struct {
	Home         string
	Popovers     []string
	Filters      []string
	HasResults   []string
	ResultsCount string
	NextPage     string
	PageNum      string
}

var blocket = Settings{
	Home:         "https://jobb.blocket.se/",
	Popovers:     []string{`iframe[id^="sp_message_iframe"]$$$button[class~=sp_choice_type_11]`},
	Filters:      []string{`div#sort-filter`, `div[data-value="sp=%s"]`},
	HasResults:   []string{`div.job-item a.header`},
	ResultsCount: `span[id="total-count"]`,
	NextPage:     `div.pagination a>i.right`,
	PageNum:      `button.blue`,
}

func main() {
	logger := zlog.MustNewZapLogger()
	xpretty.InitializeWithColor()

	selectors := blocket

	bot := wee.NewBotDefault(
		wee.WithPopovers(selectors.Popovers...),
		wee.WithPanicBy(wee.PanicByDump),
		wee.WithHighlightTimes(1),
		wee.Logger(logger),
	)

	defer bot.Cleanup()
	defer wee.Blocked()

	bot.MustOpen(selectors.Home)

	bot.MustAcceptCookies(selectors.Popovers...)

	bot.ClickWithScript(`div.title a[href$="juridik/"] label`)

	val := bot.MustElemAttr(selectors.ResultsCount)
	logger.Info("Results count", zap.String("count", val))

	bot.MustClickSequentially(selectors.Filters[0], fmt.Sprintf(selectors.Filters[1], "0"))

	for i := 0; i < 0; i++ {
		pageNum := bot.MustElemAttr(selectors.PageNum)
		logger.Info("Current page", zap.String("page", pageNum))

		results := bot.MustElemsForSelectors(selectors.HasResults)
		logger.Info("Total jobs", zap.Int("count", len(results)))

		titles := bot.MustAllElemAttrs(selectors.HasResults[0], wee.WithAttr("href"))
		logger.Info("Titles", zap.Any("titles", titles))

		elems := bot.MustElemsForSelectors(selectors.HasResults)
		res := bot.AllElementsAttrMap(elems, wee.WithAttrMap(map[string]string{
			"url":   "href",
			"title": "",
			"class": "class",
		}))
		logger.Info("Results", zap.Any("results", res))

		button, err := bot.Elem(selectors.NextPage, wee.WithTimeout(wee.NapToSec))
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Info("No more next page button found")
			bot.MustScrollToBottom()

			break
		}

		bot.MustClickElemAndWait(button)
	}

	url := bot.CurrentURL()
	logger.Info("Final URL", zap.String("url", url))
}
