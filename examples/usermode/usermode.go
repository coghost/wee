package main

import (
	"wee"

	"github.com/coghost/xlog"
	"github.com/coghost/xpretty"
	"github.com/rs/zerolog/log"
)

func main() {
	xlog.InitLogDebug()
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

	sel := bot.MustAnyElem(results, noResults)

	if sel == noResults {
		log.Warn().Msg("no results found")
		return
	}

	elems, err := bot.Elems(results)
	if err != nil {
		log.Error().Err(err).Msg("cannot get results")
	}

	for _, elem := range elems {
		xpretty.CyanPrintln(elem.MustText())
	}
}
