package wee

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

var (
	// ErrNoSelectorClicked is returned when no cookie acceptance selectors could be clicked successfully.
	ErrNoSelectorClicked = errors.New("failed to click any cookie acceptance selectors")

	// ErrPageLoadAfterCookies is returned when the page fails to load after attempting to accept cookies.
	ErrPageLoadAfterCookies = errors.New("page failed to load after accepting cookies")
)

func (b *Bot) MustClickAndWait(selector string, opts ...ElemOptionFunc) {
	b.MustClick(selector, opts...)
	b.page.MustWaitStable()
}

func (b *Bot) MustClickAll(selectors []string, opts ...ElemOptionFunc) {
	for _, ss := range selectors {
		b.MustClick(ss, opts...)
	}
}

func (b *Bot) MustClick(selector string, opts ...ElemOptionFunc) {
	defer b.LogTimeSpent(time.Now())

	b.pie(b.Click(selector, opts...))
}

// Click finds an element using the given selector and attempts to click it.
//
// This function performs the following steps:
//  1. Locates an element matching the provided CSS selector.
//  2. If found, attempts to click the element using the Bot's ClickElem method.
//
// The function uses the Bot's configured timeout settings and other properties
// for element selection and interaction behavior. It supports customization
// through optional ElemOptionFunc arguments, which can modify both the element
// selection process and the click behavior.
//
// If no element matches the selector, the function returns an ErrCannotFindSelector error.
// If an element is found but cannot be clicked (e.g., not interactable or covered
// by another element), the error from the ClickElem method is returned.
//
// Usage:
//
//	err := bot.Click("#submit-button")
//	if err != nil {
//	    // Handle error (element not found, not clickable, etc.)
//	}
//
// This method is useful for simulating user interactions in web automation tasks,
// particularly when dealing with dynamic content or complex DOM structures.
// It encapsulates the logic of finding and interacting with elements, providing
// a simple interface for common click operations.
func (b *Bot) Click(selector string, opts ...ElemOptionFunc) error {
	elem, err := b.Elem(selector, opts...)
	if err != nil {
		return err
	}

	if elem == nil {
		return ErrCannotFindSelector(selector)
	}

	return b.ClickElem(elem)
}

func (b *Bot) MustClickElemAndWait(elem *rod.Element, opts ...ElemOptionFunc) {
	b.MustClickElem(elem, opts...)
	b.page.MustWaitStable()
}

func (b *Bot) MustClickElem(elem *rod.Element, opts ...ElemOptionFunc) {
	b.pie(b.ClickElem(elem, opts...))
}

// ClickElem attempts to click the given element.
// It performs the following steps:
//  1. Optionally highlights the element.
//  2. Ensures the element is interactable (scrolls into view if necessary).
//  3. Attempts to click the element using the left mouse button.
//  4. If the click fails due to a timeout or invisibility, it falls back to clicking via JavaScript.
//
// Parameters:
//   - elem: The rod.Element to be clicked.
//   - opts: Optional ElemOptionFunc to customize the click behavior.
//
// Options:
//   - handleCoverByEsc: If true, attempts to handle covered elements by pressing Escape.
//   - highlight: If true, highlights the element before clicking.
//   - clickByScript: If true, uses JavaScript to perform the click instead of simulating a mouse click.
//
// Returns:
//   - An error if the click operation fails, nil otherwise.
func (b *Bot) ClickElem(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{handleCoverByEsc: true, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.clickByScript {
		return b.ClickElemWithScript(elem, opts...)
	}

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	err := b.MakeElemInteractable(elem, opt.handleCoverByEsc)
	if err != nil {
		return fmt.Errorf("failed to make element interactable: %w", err)
	}

	err = elem.Timeout(b.shortTimeout).Click(proto.InputMouseButtonLeft, 1)
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, &rod.InvisibleShapeError{}) {
		return b.ClickElemWithScript(elem, opts...)
	}

	return err
}

func (b *Bot) MakeElemInteractable(elem *rod.Element, byEsc bool) error {
	err := b.Interactable(elem)
	if err == nil {
		return nil
	}

	if hit := b.ClosePopovers(b.popovers...); hit != 0 {
		return nil
	}

	if errors.Is(err, &rod.CoveredError{}) && byEsc {
		return b.PressEscape(elem)
	}

	return err
}

func (b *Bot) Interactable(elem *rod.Element) error {
	vis, err := elem.Visible()
	if !vis {
		return ErrNotVisible
	}

	if err != nil {
		return fmt.Errorf("error checking visibility: %w", err)
	}

	_, err = elem.Interactable()
	if err != nil {
		scrollErr := elem.ScrollIntoView()
		if scrollErr != nil {
			return fmt.Errorf("error scrolling into view: %w", scrollErr)
		}

		_, err = elem.Interactable()
		if err != nil {
			return fmt.Errorf("element not interactable after scrolling: %w", err)
		}
	}

	return nil
}

func (b *Bot) MustPressEscape(elem *rod.Element) {
	b.pie(b.PressEscape(elem))
}

func (b *Bot) PressEscape(elem *rod.Element) error {
	return b.Press(elem, input.Escape)
}

func (b *Bot) Press(elem *rod.Element, keys ...input.Key) error {
	return elem.Timeout(b.shortTimeout).MustKeyActions().Press(keys...).Do()
}

func (b *Bot) ClickWithScript(selector string, opts ...ElemOptionFunc) error {
	elem, err := b.Elem(selector, opts...)
	if err != nil {
		return err
	}

	return b.ClickElemWithScript(elem, opts...)
}

// ClickElemWithScript attempts to click the given element using JavaScript.
// This method is particularly useful when standard clicking methods fail due to
// element positioning, overlays, or other DOM-related issues.
//
// The function performs the following steps:
//  1. Optionally highlights the element for visual feedback.
//  2. Executes a JavaScript click event on the element.
//  3. Checks if the element is still interactable after the click.
//
// Options:
//   - timeout: Duration to wait for the click operation to complete (default: MediumToSec).
//   - highlight: If true, highlights the element before clicking (default: true).
//
// Note:
//   - This method cancels the timeout after the initial JavaScript execution to allow for
//     potential page reloads or redirects triggered by the click.
//   - If the element becomes non-interactable (e.g., removed from DOM) after the click,
//     it's considered a successful operation, assuming the click caused a page change.
func (b *Bot) ClickElemWithScript(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{timeout: MediumToSec, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	_, err := elem.Timeout(time.Duration(opt.timeout)*time.Second).CancelTimeout().Eval(`(elem) => { this.click() }`, elem)
	if err != nil {
		b.logger.Error("cannot close by Eval script this.click()", zap.Error(err))
		return err
	}

	_, err = elem.Interactable()
	if errors.Is(err, &rod.ObjectNotFoundError{}) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, &rod.InvisibleShapeError{}) {
		return nil
	}

	return err
}

func (b *Bot) MustAcceptCookies(cookies ...string) {
	err := b.AcceptCookies(cookies...)
	if err != nil {
		b.logger.Error("Failed to accept cookies", zap.Error(err))
		panic(err)
	}
}

// AcceptCookies attempts to accept cookies by clicking elements matching the provided selectors.
//
// Parameters:
//   - cookies: Variadic string parameter of CSS selectors for cookie acceptance elements.
//
// Returns:
//   - error: nil if at least one selector was clicked and page loaded successfully, otherwise an error.
//
// The function tries all selectors even if some fail, logging warnings for individual failures.
// It returns an error if no selectors were clicked or if the final page load fails.
func (b *Bot) AcceptCookies(cookies ...string) error {
	if len(cookies) == 0 {
		return nil
	}

	clickedAny := false

	for _, sel := range cookies {
		err := b.Click(sel)
		if err != nil {
			b.logger.Warn("Failed to click cookie acceptance selector", zap.String("selector", sel), zap.Error(err))
		} else {
			clickedAny = true
		}
	}

	if !clickedAny {
		return ErrNoSelectorClicked
	}

	err := b.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPageLoadAfterCookies, err)
	}

	return nil
}

func (b *Bot) ClosePopovers(popovers ...string) int {
	hit := 0

	if len(popovers) == 0 {
		return 0
	}

	for _, pop := range popovers {
		n, err := b.ClosePopover(pop)
		if err == nil {
			hit += n
		} else {
			b.logger.Debug("cannot close popover", zap.Error(err))
		}
	}

	if hit != 0 {
		b.logger.Debug("closed popovers", zap.Int("count", hit))
	}

	return hit
}

// ClosePopover attempts to close popovers matching the given selector.
// It finds all elements matching the selector, checks if they're interactable,
// highlights them, and tries to click each one.
//
// Parameters:
//   - sel: CSS selector string for the popover element(s).
//
// Returns:
//   - int: Number of popovers successfully closed.
//   - error: nil if successful, otherwise an error describing the failure.
//
// The function stops and returns an error if it encounters any issues during the process.
// If some popovers are closed before an error occurs, it returns the partial count along with the error.
//
// Usage example:
//
//	closed, err := bot.ClosePopover(".modal-close-button")
//	if err != nil {
//	    log.Printf("Error closing popovers: %v", err)
//	} else {
//	    log.Printf("Successfully closed %d popovers", closed)
//	}
func (b *Bot) ClosePopover(sel string) (int, error) {
	hit := 0

	elems, err := b.Elems(sel)
	if err != nil {
		return 0, err
	}

	if len(elems) == 0 {
		return 0, nil
	}

	for _, elem := range elems {
		if _, err := elem.Interactable(); err != nil {
			// elem.Overlay("popover is not interactable")
			return 0, err
		}

		b.highlightElem(elem)

		e := elem.Click(proto.InputMouseButtonLeft, 1)
		if e != nil {
			return 0, e
		}

		hit += 1
	}

	return hit, nil
}

// MustClickSequentially clicks on a series of elements specified by selectors in order.
// It waits 500ms between clicks and ensures DOM stability after all clicks.
//
// Parameters:
//   - selectors: Variadic string parameter of CSS selectors to be clicked in sequence.
//
// This function will panic if any click operation fails.
func (b *Bot) MustClickSequentially(selectors ...string) {
	for _, sel := range selectors {
		b.MustClick(sel)
		SleepPT500Ms()
	}

	b.MustDOMStable()
}
