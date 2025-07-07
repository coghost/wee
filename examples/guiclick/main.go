package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-vgo/robotgo"
)

// Launch Chrome with remote debugging enabled
func launchChromeWithDebugging() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "--remote-debugging-port=9222")
	case "windows":
		cmd = exec.Command("C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe", "--remote-debugging-port=9222")
	case "linux":
		cmd = exec.Command("google-chrome", "--remote-debugging-port=9222")
	default:
		fmt.Println("Unsupported operating system")
		os.Exit(1)
	}

	fmt.Println("Launching Chrome with remote debugging enabled on port 9222...")

	// Start Chrome in the background
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error launching Chrome: %v\n", err)
		os.Exit(1)
	}

	// Give Chrome time to start
	time.Sleep(2 * time.Second)

	fmt.Println("Chrome launched successfully!")
}

// Click at specified coordinates with activation click first
func activateAndClick(x, y int, delay time.Duration) {
	fmt.Printf("Activating application at (%d, %d)...\n", x, y)

	// First click to activate the window
	robotgo.MoveMouse(x, y)
	robotgo.Click()

	// Wait for the specified delay
	time.Sleep(delay)

	// Second click to perform the intended action
	fmt.Printf("Performing action click at (%d, %d)...\n", x, y)
	robotgo.Click()

	fmt.Println("Activate-and-click performed successfully!")
}

// Find element in Chrome using go-rod in user mode
func findElementInUserChrome(url, selector string) (int, int, error) {
	fmt.Println("Connecting to Chrome instance...")

	// Get the websocket endpoint of the browser
	wsURL := launcher.NewUserMode().MustLaunch()

	// Connect to the browser
	browser := rod.New().ControlURL(wsURL).MustConnect()

	// Get all pages
	pages, err := browser.Pages()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get pages: %v", err)
	}

	// Find or create page with desired URL
	var page *rod.Page

	// First check if we have a page with the target URL already
	for _, p := range pages {
		pageURL := p.MustInfo().URL
		if pageURL == url {
			page = p
			fmt.Printf("Found existing page with URL: %s\n", url)
			break
		}
	}

	// If no matching page found, navigate to the URL in the first available page
	if page == nil {
		if len(pages) > 0 {
			page = pages[0]
			fmt.Printf("Navigating existing page to: %s\n", url)
			page.MustNavigate(url).MustWaitStable()
		} else {
			// Create a new page if none exist
			page = browser.MustPage(url).MustWaitStable()
			fmt.Printf("Created new page with URL: %s\n", url)
		}
	}

	// Wait for page to be stable
	page.MustWaitStable()

	fmt.Printf("Looking for element matching selector: %s\n", selector)

	// Find the element
	element, err := page.Element(selector)
	if err != nil {
		return 0, 0, fmt.Errorf("element not found: %v", err)
	}

	// Scroll element into view if needed
	err = element.ScrollIntoView()
	if err != nil {
		fmt.Println("Warning: Could not scroll element into view")
	}

	// Get the position of the element
	bounds := element.MustShape().Box()

	// Calculate the center of the element
	centerX := int(bounds.X + bounds.Width/2)
	centerY := int(bounds.Y + bounds.Height/2)

	fmt.Printf("Element found at position: (%d, %d)\n", centerX, centerY)

	return centerX, centerY, nil
}

func main() {
	fmt.Println("=== Chrome Element Finder and Clicker ===")

	// Ask whether to launch Chrome
	var launchChoice string
	fmt.Print("Do you want to launch Chrome with debugging enabled? (y/n): ")
	fmt.Scanln(&launchChoice)

	if launchChoice == "y" || launchChoice == "Y" {
		launchChromeWithDebugging()
	} else {
		fmt.Println("Make sure Chrome is already running with remote debugging enabled.")
		fmt.Println("You can launch Chrome with: chrome --remote-debugging-port=9222")
	}

	url := "https://www.google.com"
	selector := `input[name='q']`

	// Find element
	x, y, err := findElementInUserChrome(url, selector)
	if err != nil {
		fmt.Printf("Error finding element: %v\n", err)
		return
	}

	// Ask for click confirmation
	var confirm string
	fmt.Printf("Element found at (%d, %d). Proceed with click? (y/n): ", x, y)
	fmt.Scanln(&confirm)

	if confirm == "y" || confirm == "Y" {
		var delay int
		fmt.Print("Enter delay between clicks in milliseconds (default 500): ")
		fmt.Scanln(&delay)
		if delay <= 0 {
			delay = 500
		}

		activateAndClick(x, y, time.Duration(delay)*time.Millisecond)
	} else {
		fmt.Println("Click canceled.")
	}
}
