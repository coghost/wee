package main

import "github.com/coghost/wee"

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	bot.MustOpen("https://bot.sannysoft.com/")
}
