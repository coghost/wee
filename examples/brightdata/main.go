package main

import (
	"log"

	"github.com/coghost/wee"
	"github.com/k0kubun/pp/v3"
)

func main() {
	options := []wee.BotOption{}
	options = append(options, wee.WithBrowserOptions([]wee.BrowserOptionFunc{
		wee.BrowserExtensions("~/tmp/extensions/site805/superproxy"),
	}))

	bot := wee.NewBotWithOptionsOnly(options...)

	wee.BindBotLanucher(bot)

	defer bot.BlockInCleanUp()

	uri := "https://ifconfig.co/json"
	bot.Open(uri)

	ipInfo, err := wee.CurlIPFromIfconfigCO(bot)
	if err != nil {
		log.Printf("cannot get ip data info: %v", err)
	}

	pp.Println(ipInfo)
}
