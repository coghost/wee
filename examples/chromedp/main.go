package main

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"

	cu "github.com/Davincible/chromedp-undetected"
)

func main() {
	// New creates a new context for use with chromedp. With this context
	// you can use chromedp as you normally would.
	ctx, cancel, err := cu.New(cu.NewConfig(
		// Remove this if you want to see a browser window.
		// cu.WithHeadless(),
		// If the webelement is not found within 10 seconds, timeout.
		cu.WithTimeout(60 * time.Second),
	))
	if err != nil {
		panic(err)
	}
	defer cancel()

	if err := chromedp.Run(ctx,
		// Check if we pass anti-bot measures.
		chromedp.Navigate("https://www.crunchbase.com/organization/indeed/signals_and_news/timeline"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	); err != nil {
	}

	fmt.Println("Undetected!")
}
