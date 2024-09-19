package main

import (
	"github.com/coghost/wee"
	"github.com/coghost/xpretty"
	"github.com/coghost/zlog"
)

type Schema struct {
	Home         string
	CategoryLink string
	Items        []string
}

var mapper = Schema{
	Home:         "https://www.memoryexpress.com/Navigation/Group/ComputerParts",
	CategoryLink: `.l-shcn-page a[href="/Category/AdapterCards"]`,
	Items:        []string{`div.c-shca-icon-item__body-image>a`},
}

func main() {
	zlog.MustNewLoggerDebug()
	xpretty.InitializeWithColor()

	bot := wee.NewBotDefault()
	defer bot.Cleanup()

	bot.MustOpen(mapper.Home)
	bot.MustClick(mapper.CategoryLink)
	bot.MustWaitLoad()
	wsl := bot.CurrentURL() + `?InventoryType=WhileSuppliesLast`
	bot.MustOpen(wsl)

	bot.MustElemsForSelectors(mapper.Items)
}
