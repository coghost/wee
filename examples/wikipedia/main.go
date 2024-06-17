package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/coghost/zlog"
	"go.uber.org/zap"
)

func main() {
	logger := zlog.MustNewZapLogger()

	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	bot.MustOpen("https://en.wikipedia.org/wiki")

	_, err := bot.Input(`form#searchform input.cdx-text-input__input`, "golang")
	if err != nil {
		logger.Error("cannot input", zap.Error(err))
		return
	}

	// cancel the suggestions
	bot.MustClick(`form#searchform span.cdx-icon`)

	if err := bot.Click(`form#searchform button`, wee.WithTimeout(wee.ShortToSec)); err != nil {
		logger.Error("cannot submit", zap.Error(err))
		return
	}

	bot.Page().MustWaitDOMStable()

	h2Elems := bot.MustElems(`div.mw-body-content h2`)
	for _, h2 := range h2Elems {
		txt := bot.MustElemAttr(`span[id]`, wee.WithRoot(h2))
		link := bot.MustElemAttr(`a`, wee.WithAttr("href"), wee.WithRoot(h2))

		fmt.Printf("%s\n  - %s\n", txt, link)
	}
}
