package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/go-vgo/robotgo"
	"github.com/k0kubun/pp/v3"
)

func main() {
	x, y := 0, 0

	for {
		x, y = robotgo.Location()
		pp.Println("location: ", x, y)

		b := wee.Confirm("is this OK?")
		if b {
			break
		}
	}

	activateAndClick(x, y)
}

func getPosition() (int, int) {
	var x, y int

	fmt.Print("Enter x coordinate: ")
	fmt.Scanln(&x)

	fmt.Print("Enter y coordinate: ")
	fmt.Scanln(&y)

	return x, y
}

// Click at specified coordinates with activation click first
func activateAndClick(x, y int) {
	fmt.Printf("Activating application at (%d, %d)...\n", x, y)

	// First click to activate the window
	robotgo.Move(x, y)
	robotgo.Click("left", true)

	// Wait for the specified delay
	robotgo.Sleep(3)

	// Second click to perform the intended action
	fmt.Printf("Performing action click at (%d, %d)...\n", x, y)
	robotgo.Move(x, y)
	robotgo.Click("left", true)

	fmt.Println("Activate-and-click performed successfully!")
}
