package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

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

func getCurrentPositionAndClick(activateMode bool) {
	fmt.Println("Capturing current cursor position in 3 seconds...")
	fmt.Println("Move your cursor to the desired location...")
	for i := 3; i > 0; i-- {
		fmt.Printf("%d...\n", i)
		time.Sleep(1 * time.Second)
	}

	// Get the current mouse position
	x, y := robotgo.GetMousePos()

	fmt.Printf("Cursor position captured: (%d, %d)\n", x, y)

	if activateMode {
		// Perform activate-then-click with a 500ms delay
		activateAndClick(x, y, 500*time.Millisecond)
	} else {
		// Simple single click
		fmt.Printf("Clicking at position (%d, %d)...\n", x, y)
		robotgo.Click()
		fmt.Println("Click performed successfully!")
	}
}

func clickAtSpecifiedPosition(activateMode bool) {
	var x, y int

	fmt.Print("Enter x coordinate: ")
	fmt.Scanln(&x)

	fmt.Print("Enter y coordinate: ")
	fmt.Scanln(&y)

	if activateMode {
		// Perform activate-then-click with a 500ms delay
		activateAndClick(x, y, 500*time.Millisecond)
	} else {
		// Simple single click
		fmt.Printf("Clicking at position (%d, %d)...\n", x, y)
		robotgo.MoveMouse(x, y)
		robotgo.Click()
		fmt.Println("Click performed successfully!")
	}
}

func performCustomClick() {
	var x, y int
	var numClicks int
	var delayMs int

	fmt.Print("Enter x coordinate: ")
	fmt.Scanln(&x)

	fmt.Print("Enter y coordinate: ")
	fmt.Scanln(&y)

	fmt.Print("Number of clicks (1-10): ")
	fmt.Scanln(&numClicks)
	if numClicks < 1 {
		numClicks = 1
	} else if numClicks > 10 {
		numClicks = 10 // Safety limit
	}

	fmt.Print("Delay between clicks in milliseconds (0-5000): ")
	fmt.Scanln(&delayMs)
	if delayMs < 0 {
		delayMs = 0
	} else if delayMs > 5000 {
		delayMs = 5000 // Safety limit
	}

	delay := time.Duration(delayMs) * time.Millisecond

	// Move to position
	robotgo.MoveMouse(x, y)

	// Perform specified number of clicks with delay
	for i := 0; i < numClicks; i++ {
		fmt.Printf("Click %d/%d at position (%d, %d)...\n", i+1, numClicks, x, y)
		robotgo.Click()

		if i < numClicks-1 && delayMs > 0 {
			time.Sleep(delay)
		}
	}

	fmt.Println("Custom click sequence performed successfully!")
}

func getScreenInfo() {
	width, height := robotgo.GetScreenSize()
	fmt.Printf("Screen size: %dx%d pixels\n", width, height)

	x, y := robotgo.GetMousePos()
	fmt.Printf("Current mouse position: (%d, %d)\n", x, y)
}

func main() {
	fmt.Println("=== Go Mouse Automation Tool ===")
	getScreenInfo()
	fmt.Println("\nChoose an option:")
	fmt.Println("1. Capture current cursor position and single-click")
	fmt.Println("2. Capture current position with activate-then-click")
	fmt.Println("3. Enter coordinates for single-click")
	fmt.Println("4. Enter coordinates with activate-then-click")
	fmt.Println("5. Custom click sequence (multiple clicks with delay)")

	var choice string
	fmt.Print("Enter your choice (1-5): ")
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		getCurrentPositionAndClick(false)
	case "2":
		getCurrentPositionAndClick(true)
	case "3":
		clickAtSpecifiedPosition(false)
	case "4":
		clickAtSpecifiedPosition(true)
	case "5":
		performCustomClick()
	default:
		fmt.Println("Invalid choice")
	}
}
