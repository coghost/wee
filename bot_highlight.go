package wee

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

const (
	_zoomIn = `zoom: 1.25;-moz-transform: scale(1.25);`
	_style  = `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	_style1 = `box-shadow: 0 0 10px rgba(255,125,0,1), 0 0 20px 5px rgba(255,175,0,0.8), 0 0 30px 15px rgba(255,225,0,0.5);`
)

// FocusAndHighlight focuses on an element and optionally highlights it.
//
// This method provides visual emphasis on a specified web element, either by focusing
// on it or by scrolling to it and applying a highlight effect. The behavior is determined
// by the Bot's configuration.
//
// Parameters:
//   - elem: A pointer to the rod.Element to focus on or highlight.
//
// Behavior:
//  1. If Bot.highlightTimes == 0:
//     - The method only focuses on the element using elem.Focus().
//     - No scrolling or highlighting is performed.
//  2. If Bot.highlightTimes > 0:
//     - The method scrolls to the element using Bot.ScrollToElement(elem).
//     - Then it calls Bot.HighlightElem(elem) to apply a visual highlight effect.
//
// The HighlightElem method:
//   - Applies a pulsating highlight effect to the element.
//   - The number of pulses is determined by Bot.highlightTimes.
//   - Uses a predefined style (_style1) for the highlight effect.
//   - Runs asynchronously in a separate goroutine.
//
// Use cases:
//   - Automated testing: To visually verify which elements are being interacted with.
//   - Debugging: To easily spot the elements being processed during script execution.
//   - Demonstrations: To create visual cues when showcasing web automation.
//
// Note:
//   - This method does not return any value or error.
//   - The highlight effect, if applied, does not block the execution of subsequent code.
//
// Example usage:
//
//	bot := &Bot{highlightTimes: 3}
//	element := page.MustElement("#some-id")
//	bot.FocusAndHighlight(element)
//
// In this example, the element will be scrolled into view and highlighted 3 times.
func (b *Bot) FocusAndHighlight(elem *rod.Element) {
	if b.highlightTimes == 0 {
		_ = elem.Focus()
	} else {
		b.ScrollToElement(elem)
		b.highlightElem(elem)
	}
}

// highlightElem applies a visual highlight effect to a specified web element.
//
// This is an unexported method of the Bot struct that initiates the highlighting
// process for a given element.
//
// Parameters:
//   - elem: A pointer to the rod.Element to be highlighted.
//
// Behavior:
//  1. If Bot.highlightTimes == 0:
//     - The method returns immediately without any effect.
//  2. If Bot.highlightTimes > 0:
//     - A highlight effect is applied to the element asynchronously.
//
// Highlight Effect Details:
//   - Duration visible (show): 0.333 seconds
//   - Duration hidden (hide): 0.2 seconds
//   - Style: Defined by _style1 constant (a glowing box-shadow effect)
//   - Number of pulses: Determined by Bot.highlightTimes
//
// Implementation:
//   - The actual highlighting is performed by the Highlight method.
//   - It's called in a separate goroutine to avoid blocking execution.
//
// Note:
//   - This method is intended for internal use within the Bot struct.
//   - The highlight effect runs asynchronously and doesn't block execution.
//
// The method is typically called by FocusAndHighlight when highlighting is enabled.
func (b *Bot) highlightElem(elem *rod.Element) {
	if b.highlightTimes == 0 {
		return
	}

	show, hide := 0.333, 0.2

	go b.Highlight(elem, show, hide, _style1, 0)
}

// Highlight applies a pulsating visual effect to a specified web element.
//
// This method creates a series of highlight pulses on the given element by
// alternating between a custom style and the element's original style.
//
// Parameters:
//   - elem: A pointer to the rod.Element to be highlighted.
//   - show: Duration (in seconds) for which the highlight is visible.
//   - hide: Duration (in seconds) for which the highlight is hidden.
//   - style: CSS style string to be applied for the highlight effect.
//   - count: Number of times to repeat the highlight effect. If 0, uses Bot.highlightTimes.
//
// Returns:
//   - float64: The total duration (in seconds) taken to complete the highlight effect.
//
// Behavior:
//  1. If Bot.highlightTimes == 0 or elem is nil, the method returns immediately.
//  2. Retrieves the original style of the element.
//  3. Applies the highlight effect 'count' number of times:
//     - Sets the element's style to the provided highlight style.
//     - Waits for 'show' duration (with slight randomization).
//     - Reverts the element's style to its original state.
//     - Waits for 'hide' duration (with slight randomization).
//  4. Measures and returns the total time taken for the entire process.
//
// Notes:
//   - The method uses JavaScript evaluation to modify the element's style.
//   - There's a small randomization (±0.05s) added to show/hide durations for a more natural effect.
//   - This method is synchronous and will block until all highlight pulses are completed.
//
// Example usage:
//
//	bot := &Bot{highlightTimes: 3}
//	element := page.MustElement("#some-id")
//	duration := bot.Highlight(element, 0.5, 0.3, "border: 2px solid red;", 0)
//	fmt.Printf("Highlighting took %.2f seconds\n", duration)
func (b *Bot) Highlight(elem *rod.Element, show, hide float64, style string, count int) float64 {
	start := time.Now()

	if b.highlightTimes == 0 {
		return 0
	}

	if elem == nil {
		return 0
	}

	origStyle := ""

	ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	if err == nil {
		origStyle = ob.Value.String()
	}

	// origStyle := elem.MustEval(`e => {return this.getAttribute("style")}`).String()
	// style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	base := 0.05

	if count == 0 {
		count = b.highlightTimes
	}

	for i := 0; i < count; i++ {
		script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, style)
		_, _ = elem.Eval(script)

		RandSleep(show-base, show+base)

		script = fmt.Sprintf(`() => this.setAttribute("style", "%s");`, origStyle)
		_, _ = elem.Eval(script)

		RandSleep(hide-base, hide+base)
	}

	cost := time.Since(start).Seconds()

	return cost
}

// MarkElems applies a temporary visual marker to multiple web elements concurrently.
//
// Parameters:
//   - timeout: Duration for which the marker style should be applied.
//   - elems: Variadic parameter of rod.Element pointers to be marked.
//
// Behavior:
//   - Applies a purple box-shadow style to each element for the specified duration.
//   - Marking is done concurrently for each element using goroutines.
//   - Uses MarkElem internally to apply the style to each element.
//
// Note:
//   - This method returns immediately, with marking occurring asynchronously.
//   - The default style is a purple box-shadow.
func (b *Bot) MarkElems(timeout time.Duration, elems ...*rod.Element) {
	style := `box-shadow: 0px 0px 0px 3px rgba(148,0,211,1);`
	for _, elem := range elems {
		go b.MarkElem(timeout, elem, style)
	}
}

// MarkElem applies a temporary visual marker to a specified web element.
//
// This method adds a visual style to an element for a specified duration, then
// reverts the element to its original style. It's useful for temporarily
// highlighting or marking elements during automated browsing or testing.
//
// Parameters:
//   - timeout: Duration for which the marker style should be applied.
//   - elem: A pointer to the rod.Element to be marked.
//   - style: CSS style string to be applied as the marker. If empty, a default style is used.
//
// Behavior:
//  1. Captures the original style of the element (currently commented out).
//  2. Applies the specified style (or default if not provided) to the element.
//  3. Waits for the specified timeout duration.
//  4. Reverts the element's style to its original state (deferred operation).
//
// Default Style:
//
//	If no style is provided, it uses the _style constant:
//	`box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px;
//	 transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
//
// Notes:
//   - This method is synchronous and will block for the duration of the timeout.
//   - The original style restoration is deferred, ensuring it occurs even if there's a panic.
//   - Currently, the original style capture is commented out, which means the element
//     will revert to having no inline style after the timeout.
//
// Example usage:
//
//	bot := &Bot{}
//	element := page.MustElement("#some-id")
//	bot.MarkElem(5*time.Second, element, "border: 3px solid blue;")
//
// This will mark the element with a blue border for 5 seconds.
func (b *Bot) MarkElem(timeout time.Duration, elem *rod.Element, style string) {
	origStyle := ""

	// ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	// if err == nil {
	// 	origStyle = ob.Value.String()
	// }

	defer func() {
		script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, origStyle)
		_, _ = elem.Eval(script)
	}()

	if style == "" {
		style = _style
	}

	script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, style)
	_, _ = elem.Eval(script)

	time.Sleep(timeout)
}
