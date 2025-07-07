package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	lch := launcher.NewUserMode()
	lch = lch.Bin("/Applications/Google Chrome Dev.app/Contents/MacOS/Google Chrome Dev")
	lch = lch.Headless(false)
	wsURL := lch.MustLaunch()

	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()
	browser.MustPage("https://www.indeed.com/")

	fmt.Println("done")
	wee.Confirm("continue?")
}
