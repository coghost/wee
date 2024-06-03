package main

import (
	"github.com/coghost/wee"
)

const (
	home   = "https://getbootstrap.com/docs/4.0/components/modal/"
	button = `button[data-whatever="@getbootstrap"]`
)

func main() {
	bot := wee.NewBotDefault()

	defer bot.Cleanup()
	defer wee.QuitOnTimeout(2)

	bot.MustOpen(home)
	elem, _ := bot.Elem(button)
	bot.ClickElem(elem)
}
