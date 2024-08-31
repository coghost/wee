package wee

import (
	"fmt"
	"os"
	"time"

	"github.com/coghost/xdtm"
	"github.com/coghost/xpretty"
	"github.com/pterm/pterm"
	"github.com/spf13/cast"
)

// Blocked prompts the user to decide whether to quit or continue indefinitely.
// It calls QuitOnTimeout with a negative value to trigger the interactive prompt.
func Blocked() {
	QuitOnTimeout(-1)
}

// QuitOnTimeout waits for a specified number of seconds before quitting the program.
//   - If no argument is provided, it uses a default timeout of 3 seconds.
//   - If the timeout is 0, it returns immediately without any action.
//   - If the timeout is negative, it prompts the user to quit or continue.
//   - Otherwise, it displays a spinner for the specified duration before quitting.
//
// Parameters:
//   - args: Variadic parameter for timeout in seconds. If not provided, default is used.
func QuitOnTimeout(args ...int) {
	dft := 3

	sec := FirstOrDefault(dft, args...)
	if sec == 0 {
		return
	}

	if sec < 0 {
		QuitIfY()
		return
	}

	SleepWithSpin(sec)
}

// SleepWithSpin displays a spinner for a specified number of seconds with an optional message.
// It shows a visual indicator of waiting and then exits after the specified duration.
//
// Parameters:
//   - timeInSeconds: The duration to display the spinner, in seconds.
//   - args: Optional. A custom message to display. If not provided, a default message is used.
func SleepWithSpin(timeInSeconds int, args ...string) {
	msg := fmt.Sprintf("Wait for %d seconds before quitting ...", timeInSeconds)
	if len(args) > 0 {
		msg = args[0]
	}

	spinnerInfo, _ := pterm.DefaultSpinner.Start(xpretty.Yellow(msg))

	time.Sleep(time.Second * time.Duration(timeInSeconds))
	spinnerInfo.Info()
}

// Confirm displays an interactive confirmation prompt and returns the user's choice as a boolean.
// It uses pterm's DefaultInteractiveConfirm to show the prompt.
//
// Parameters:
//   - msg: Optional. A custom message for the confirmation prompt. If not provided, a default message is used.
//
// Returns:
//   - bool: true if the user confirms, false otherwise.
func Confirm(msg ...string) bool {
	b, _ := pterm.DefaultInteractiveConfirm.Show(msg...)
	return b
}

// TermScanf prompts the user for text input and returns the input as a string.
// It uses pterm's DefaultInteractiveTextInput to display the prompt and capture user input.
//
// Parameters:
//   - msg: Optional. A custom message for the input prompt. If not provided, a default prompt is used.
//
// Returns:
//   - string: The user's input as a string.
func TermScanf(msg ...string) string {
	result, _ := pterm.DefaultInteractiveTextInput.Show(msg...)
	return result
}

// TermScanfInt prompts the user for numeric input and returns the input as an integer.
// It uses pterm's DefaultInteractiveTextInput to display the prompt and capture user input,
// then converts the input to an integer using cast.ToInt.
//
// Parameters:
//   - msg: Optional. A custom message for the input prompt. If not provided, a default prompt is used.
//
// Returns:
//   - int: The user's input converted to an integer.
func TermScanfInt(msg ...string) int {
	result, _ := pterm.DefaultInteractiveTextInput.Show(msg...)
	return cast.ToInt(result)
}

// QuitIfY prompts the user to quit or continue the program.
// It displays an interactive confirmation dialog with the current date,
// asking if the user wants to quit (Y) or continue (N).
// If the user chooses to quit, the program exits immediately.
// The default option is set to continue (N).
func QuitIfY() {
	pd := pterm.DefaultInteractiveConfirm
	pd.DefaultValue = false
	msg := fmt.Sprintf("%s: Wanna quit(Y) or continue(N)?", xdtm.CarbonNow().ToShortDateString())

	if b, _ := pd.Show(msg); b {
		os.Exit(0)
	}
}
