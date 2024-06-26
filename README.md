# wee

Wee is based on go-rod, and provides clean interface to write crawler.

## Examples

```go
package main

import (
	"fmt"

	"github.com/coghost/wee"
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
	bot.MustInput(input, "golang", wee.WithSubmit(true))

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
```

See [examples folder](https://github.com/coghost/wee/tree/main/examples) for more detailed examples.

## Issues

### panic: context canceled

context canceled issue usually happend when elem is bound with Timeout on `bot.Elem(xxx)`, you can use

```go
elem.CancelTimeout().Timeout(xxx) to rewrite previous value.
```
