package main

import (
	"wee"

	"github.com/rs/zerolog/log"
)

const (
	home  = "http://www.spiderbuf.cn/"
	input = `input[id="kw"]`
	cands = `div#content_left div.result-op  h3>a`
	enter = `button.btn`
	h1p   = `h1>p`
)

func main() {
	bot := wee.NewBotWithDefault()

	defer bot.Cleanup()
	defer bot.QuitOnTimeout(3)

	bot.MustOpen(home)

	elem, _ := bot.GetElem(h1p)

	err := bot.ClickElement(elem)
	if err != nil {
		log.Error().Err(err).Msg("cannot click elem")
	}
}
