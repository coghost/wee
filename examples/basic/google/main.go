package main

import (
	"fmt"

	"github.com/coghost/wee"
)

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	const (
		url     = "https://www.google.com/"
		input   = `textarea[title]`
		keyword = "golang"
		items   = `div[jsaction] span[jsaction]>a[jsname]`
		noItems = `div.this_should_not_existed`
	)

	bot.MustOpen(url)
	bot.MustInput(input, keyword, wee.WithSubmit(true))

	got := bot.MustAnyElem([]string{items, noItems})
	if got == noItems {
		fmt.Printf("%s has no results\n", keyword)
		return
	}

	elems, err := bot.Elems(items)
	if err != nil {
		fmt.Printf("get items failed: %v\n", err)
	}

	for _, elem := range elems {
		fmt.Printf("%s - %s\n", elem.MustText(), *elem.MustAttribute("href"))
	}
}
