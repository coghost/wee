package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/k0kubun/pp/v3"
)

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	const (
		url     = `http://www.baidu.com`
		input   = `input[id="kw"]`
		keyword = "golang"
		items   = `div.result[id] h3>a`
		noItems = `div.this_should_not_existed`
	)

	bot.MustOpen(url)
	// bot.MustInput(input, "golang", wee.WithSubmit(true))

	got1 := bot.MustEval(`() => {
  function extractCards() {
    let cards = { a: 'OK', b: 1 }
    return cards
  }

  window.extractCards = extractCards
  return 'DONE'
}
`)

	pp.Println(got1)
	wee.Confirm("wanna continue?")

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
