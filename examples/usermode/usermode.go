package main

import (
	"github.com/coghost/wee"
	"github.com/coghost/xpretty"
	"github.com/coghost/zlog"
)

func main() {
	log := zlog.MustNewZapLogger()
	xpretty.InitializeWithColor()

	bot := wee.NewBotUserMode()

	defer bot.Cleanup()

	bot.MustOpen("http://www.baidu.com")

	const (
		input     = `input[id="kw"]`
		results   = `div#content_left div.result-op h3>a`
		noResults = `div.thiscannotbeexisted`
	)

	bot.MustInput(input, "python", wee.WithSubmit(true))

	sel := bot.MustAnyElem([]string{results, noResults})

	if sel == noResults {
		log.Info("no results found")
		return
	}

	elems, err := bot.Elems(results)
	if err != nil {
		log.Sugar().Infof("cannot get results: %v", err)
	}

	for _, elem := range elems {
		xpretty.CyanPrintln(elem.MustText())
	}
}
