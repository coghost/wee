package main

import (
	"fmt"
	"time"

	"github.com/coghost/wee"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-vgo/robotgo"
	"github.com/k0kubun/pp/v3"
)

const cbURL = "https://www.crunchbase.com/"

func main() {
	// options := []wee.BotOption{}
	// options = append(options, wee.UserMode(true))
	// options = append(options, wee.WithBrowserOptions(
	// 	[]wee.BrowserOptionFunc{
	// 		wee.LaunchLeakless(true),
	// 		wee.BrowserFlags("disable-default-apps", "no-first-run"),
	// 	}))
	// bot := wee.NewBotWithOptionsOnly(options...)
	// wee.BindBotLanucher(bot)

	bot := wee.NewBotUserMode()
	bot.Open(cbURL)

	defer wee.Blocked()

	elem, _ := bot.Elem(`p.h2@@@Verify you are human by completing the action below.`, wee.WithTimeout(wee.MediumToSec))
	if elem == nil {
		pp.Println("no iframe found")
		return
	}

	pp.Println("got iframe")

	x, y := getPosition(1, 5)

	// Click the element using robotgo
	activateAndClick(x, y, 500*time.Millisecond)

	pp.Println("done")
}

func getPosition(mode int, timeout int) (int, int) {
	if mode == 1 {
		fmt.Println("Capturing current cursor position in 3 seconds...")
		fmt.Println("Move your cursor to the desired location...")
		for i := timeout; i > 0; i-- {
			fmt.Printf("%d...\n", i)
			time.Sleep(1 * time.Second)
		}

		x, y := robotgo.Location()

		return x, y
	}

	var x, y int

	fmt.Print("Enter x coordinate: ")
	fmt.Scanln(&x)

	fmt.Print("Enter y coordinate: ")
	fmt.Scanln(&y)

	return x, y
}

// Click at specified coordinates with activation click first
func activateAndClick(x, y int, delay time.Duration) {
	fmt.Printf("Activating application at (%d, %d)...\n", x, y)

	// First click to activate the window
	robotgo.Move(x, y)
	robotgo.Click()

	// Wait for the specified delay
	time.Sleep(delay)

	// Second click to perform the intended action
	fmt.Printf("Performing action click at (%d, %d)...\n", x, y)
	robotgo.Click()

	fmt.Println("Activate-and-click performed successfully!")
}

func rodLib() {
	wsURL := launcher.NewUserMode().Leakless(true).
		UserDataDir("/tmp/t").
		Set("disable-default-apps").
		Set("no-first-run").MustLaunch()

	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	page := browser.MustPage(cbURL)
	page.MustWaitDOMStable()
}
