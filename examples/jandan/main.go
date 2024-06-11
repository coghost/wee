package main

import (
	"github.com/coghost/wee"
	"github.com/coghost/xlog"
)

func main() {
	xlog.InitLogDebug()

	bot := wee.NewBotDefault()
	defer bot.Cleanup()
	defer wee.Blocked()

	bot.DisableImages()
	bot.MustOpen("https://jandan.net/")
}
