package main

import (
	"wee"
)

const (
	home   = "https://getbootstrap.com/docs/4.0/components/modal/"
	button = `button[data-whatever="@getbootstrap"]`
)

func main() {
	bot := wee.NewBotWithDefault()

	defer bot.Cleanup()
	defer bot.QuitOnTimeout(2)

	bot.MustOpen(home)
	elem, _ := bot.GetElem(button)
	bot.ClickElement(elem)
}
