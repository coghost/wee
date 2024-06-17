package main

import (
	"github.com/coghost/wee"
)

func main() {
	bot := wee.NewBotDefault()
	defer bot.Cleanup()
	defer wee.Blocked()

	bot.DisableImages()
	bot.MustOpen("https://jandan.net/")
}
