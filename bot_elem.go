package wee

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/coghost/xpretty"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/gookit/goutil/sysutil"
	"go.uber.org/zap"
)

const (
	_logIfTimeout = 2.0
	_iframeLen    = 2
)

const (
	_textPartialMatch  = 2
	_textExactMatchLen = 3
)

var (
	ErrSelectorEmpty   = errors.New("selector is empty")
	ErrNotInteractable = errors.New("elem not interactable")
	ErrNotVisible      = errors.New("elem is not visible")

	ErrInvalidByTextFormat = fmt.Errorf("invalid selector format")
)

func (b *Bot) MustElem(selector string, opts ...ElemOptionFunc) *rod.Element {
	elem, err := b.Elem(selector, opts...)
	b.pie(err)

	return elem
}

// Elem gets an element by selector, with performance logging and error handling.
//
// This function is a wrapper around the internal getElem method, providing additional
// functionality such as performance timing and debug logging.
//
// Elem supports various selector strategies including:
//   - Standard CSS selectors
//   - Iframe-based selection (using IFrameSep)
//   - Text content-based selection (using SEP)
//   - Index-based selection for multiple matching elements
//
// The function measures the time taken to find the element and logs a debug message
// if it exceeds _logIfTimeout seconds.
//
// Parameters:
//   - selector: string to identify the element. Can be a CSS selector, iframe selector, or text content.
//   - opts: variadic ElemOptionFunc arguments to customize selection behavior.
//
// Returns:
//   - *rod.Element: pointer to the found element.
//   - error: nil if successful, otherwise an error describing why the element couldn't be found.
//
// The function respects the following options (passed via opts):
//   - WithRoot: to specify a root element for the search
//   - WithTimeout: to set a custom timeout for waiting for the element
//   - WithIndex: to select a specific element when multiple matches are found
//   - WithIframe: to specify an iframe context for the search
//
// Example usage:
//
//	elem, err := b.Elem("div.class", WithTimeout(5), WithIndex(2))
//	if err != nil {
//	    // Handle error
//	}
//	// Use elem
//
// Note: This function is part of the Bot struct and is designed to be used as a public API.
// It includes error handling and performance logging, making it suitable for external use.
func (b *Bot) Elem(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	start := time.Now()

	elem, err := b.getElem(selector, opts...)
	if err != nil {
		return nil, err
	}

	if cost := time.Since(start).Seconds(); cost > _logIfTimeout {
		b.logger.Debug("get elem", zap.Float64("cost", cost), zap.String("selector", selector))
	}

	return elem, nil
}

// ElemByIndex retrieves an element at a specific index from a list of elements matching the selector.
//
// Parameters:
//   - selector: CSS selector string
//   - index: desired element index (supports negative indexing)
//
// Returns:
//   - *rod.Element: element at specified index, or nil if index out of range
//   - error: if elements can't be found or other issues occur
//
// Note: This function doesn't wait for elements to appear.
func (b *Bot) ElemByIndex(selector string, index int) (*rod.Element, error) {
	elems, err := b.Elems(selector)
	if err != nil {
		return nil, err
	}

	index = NormalizeSliceIndex(len(elems), index)
	if index < 0 {
		return nil, nil
	}

	return elems[index], nil
}

// ElemByText finds an element by its text content using a combination of CSS selector and text matching.
//
// This function provides a powerful way to select elements based on both their structural properties
// (via CSS selectors) and their text content. It supports two types of text matching: partial (contains)
// and exact match, with an option for case-insensitive matching.
//
// Selector format:
//   - Partial match: "cssSelector@@@textContent"
//     Finds elements matching the CSS selector that contain the specified text.
//   - Exact match: "cssSelector@@@---@@@textContent"
//     Finds elements matching the CSS selector with text that exactly matches the specified content.
//
// Parameters:
//   - selector: A string combining CSS selector and text content, separated by SEP ("@@@").
//     The CSS selector comes first, followed by one or two SEP, and then the text to match.
//   - opts: Optional variadic ElemOptionFunc arguments to customize the search behavior.
//
// Returns:
//   - *rod.Element: A pointer to the found element.
//   - error: An error if the element couldn't be found or if any other issues occurred during the search.
//
// Behavior:
//  1. Parses the selector string to extract CSS selector and text content.
//  2. For exact match (three parts after splitting), constructs a regex pattern.
//  3. Uses Rod's ElementR method to perform the element search with text matching.
//  4. Applies specified timeout to the search operation.
//
// Options:
//   - WithRoot(root): Specifies a root element to search within, limiting the scope of the search.
//   - WithTimeout(timeout): Sets a duration to wait for the element to appear.
//   - WithCaseInsensitive(): Enables case-insensitive matching for exact match selectors.
//
// Example usage:
//
//	// Partial match (case-sensitive)
//	elem, err := b.ElemByText("button@@@Submit")
//
//	// Exact match (case-sensitive)
//	elem, err := b.ElemByText("div@@@---@@@Exact Text")
//
//	// Exact match (case-insensitive)
//	elem, err := b.ElemByText("span@@@---@@@Case Insensitive", WithCaseInsensitive())
//
//	// With custom timeout
//	elem, err := b.ElemByText("p@@@Content", WithTimeout(10))
//
//	// Search within a specific root element
//	rootElem, _ := b.Elem("#root-container")
//	elem, err := b.ElemByText("li@@@Item", WithRoot(rootElem))
//
// Internal workings:
//   - For partial match, it uses Rod's built-in contains text matching.
//   - For exact match, it constructs a regex pattern: /^textContent$/
//     With case-insensitive option, it adds the 'i' flag: /^textContent$/i
//
// Error handling:
//   - Returns an error if the element is not found within the specified timeout.
//   - Propagates any errors encountered during the Rod operations.
//
// Note:
//   - This function is particularly useful for selecting elements in dynamic content
//     where IDs or classes might change, but text content remains consistent.
//   - The case-insensitive option only applies to exact match selectors.
//   - When using WithRoot, ensure the root element is already present in the DOM.
//
// See also:
//   - Elem: For standard CSS selector-based element selection.
//   - ElemsByText: For selecting multiple elements by text content.
func (b *Bot) ElemByText(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	opt := ElemOptions{root: b.root, timeout: ShortToSec}
	bindElemOptions(&opt, opts...)

	arr := strings.Split(selector, SEP)
	if len(arr) < _textPartialMatch {
		return nil, ErrInvalidByTextFormat
	}

	txt := arr[len(arr)-1]

	if len(arr) == _textExactMatchLen {
		m := "/^%s$/"
		if opt.caseInsensitive {
			m += "i"
		}

		txt = fmt.Sprintf(m, txt)
	}

	var (
		elem *rod.Element
		err  error
	)

	dur := time.Duration(opt.timeout) * time.Second

	if opt.root != nil {
		elem, err = opt.root.Timeout(dur).ElementR(arr[0], txt)
	} else {
		elem, err = b.page.Timeout(dur).ElementR(arr[0], txt)
	}

	return elem, err
}

// ElemsByText finds all elements by their text content using a combination of CSS selector and text matching.
//
// This function retrieves multiple elements that match both a CSS selector and
// contain specific text content. It's useful for scenarios where you need to find
// all instances of elements with certain text.
//
// Selector format:
//   - "cssSelector@@@textContent"
//     The CSS selector and text content are separated by SEP ("@@@").
//
// Parameters:
//   - selector: A string combining CSS selector and text content, separated by SEP.
//     Format: "cssSelector@@@textContent"
//   - opts: Optional variadic ElemOptionFunc arguments to customize the search behavior.
//
// Returns:
//   - []*rod.Element: A slice of pointers to all found elements. May be empty if no elements are found.
//   - error: An error if elements couldn't be found or if any other issues occurred during the search.
//
// Behavior:
//  1. Splits the selector by SEP to separate CSS selector and text content.
//  2. Uses the internal elems method to find all elements matching the CSS selector.
//  3. Filters these elements to find all that contain the specified text content.
//
// Error handling:
//   - Returns an error if the initial elems call fails.
//   - Returns an empty slice and nil error if no elements match both criteria.
//
// Options:
//
//	Options are passed through to both the elems call and text matching. Notable options include:
//	- WithTimeout: Sets a duration to wait for elements to appear.
//	- WithRoot: Specifies a root element to search within, limiting the scope of the search.
//	- WithCaseInsensitive: Enables case-insensitive text matching.
//
// Example usage:
//
//	// Find all buttons containing "Submit"
//	elems, err := b.ElemsByText("button@@@Submit")
//
//	// Find all list items with specific text, case-insensitive
//	elems, err := b.ElemsByText("li@@@Item Text", WithCaseInsensitive())
//
// Note:
//   - This function is useful for finding multiple elements in dynamic content where both
//     structure (CSS) and content (text) need to match.
//
// See also:
//   - ElemByText: For selecting a single element by text content.
//   - Elems: For selecting multiple elements by CSS selector only.
func (b *Bot) ElemsByText(selector string, opts ...ElemOptionFunc) ([]*rod.Element, error) {
	arr := strings.Split(selector, SEP)
	if len(arr) < _textPartialMatch {
		return nil, ErrInvalidByTextFormat
	}

	cssSelector := arr[0]
	textContent := strings.Join(arr[1:], SEP)

	allElems, err := b.elems(cssSelector, opts...)
	if err != nil {
		return nil, fmt.Errorf("ElemsByText failed to get elements: %w", err)
	}

	var matchingElems []*rod.Element

	opt := ElemOptions{}
	bindElemOptions(&opt, opts...)

	for _, elem := range allElems {
		text, err := elem.Text()
		if err != nil {
			continue // Skip elements where text can't be retrieved
		}

		if opt.caseInsensitive {
			if strings.Contains(strings.ToLower(text), strings.ToLower(textContent)) {
				matchingElems = append(matchingElems, elem)
			}
		} else {
			if strings.Contains(text, textContent) {
				matchingElems = append(matchingElems, elem)
			}
		}
	}

	return matchingElems, nil
}

func (b *Bot) MustElemsForSelectors(selectors []string, opts ...ElemOptionFunc) []*rod.Element {
	var elems []*rod.Element
	for _, s := range selectors {
		elems = append(elems, b.MustElems(s)...)
	}

	return elems
}

func (b *Bot) MustElems(selector string, opts ...ElemOptionFunc) []*rod.Element {
	elems, err := b.Elems(selector, opts...)
	b.pie(err)

	return elems
}

// Elems retrieves all elements matching the given selector immediately.
//
// This function is designed to find multiple elements on the page that match
// the provided selector. It has specific behavior for different types of selectors
// and includes built-in error checking.
//
// Behavior:
//  1. If the selector is empty, it returns an ErrSelectorEmpty error.
//  2. If the selector contains the SEP ("@@@"), it calls ElemsByText for text-based selection.
//  3. Otherwise, it calls the internal elems method for standard CSS selector-based selection.
//
// The function does not wait for elements to appear and returns the current state of the page immediately.
//
// Parameters:
//   - selector: A string representing the selector to locate the elements.
//   - For standard selection: CSS selector (e.g., "div.class", "#id")
//   - For text-based selection: Selector with SEP (e.g., "div@@@text")
//   - opts: Optional ElemOptionFunc arguments to customize the element search behavior.
//     These are passed through to the underlying selection methods.
//
// Returns:
//   - []*rod.Element: A slice of pointers to the found elements. May be empty if no elements are found.
//   - error: An error if elements couldn't be found or if any other issues occurred.
//     Specifically returns ErrSelectorEmpty if the selector is an empty string.
//
// Error handling:
//   - Returns ErrSelectorEmpty if the selector is an empty string.
//   - Other errors may be returned from the underlying ElemsByText or elems methods.
//
// Usage examples:
//
//  1. Standard CSS selector:
//     elems, err := b.Elems("div.class")
//
//  2. Text-based selector:
//     elems, err := b.Elems("button@@@Submit")
//
// Note:
//   - This function is part of the Bot struct and is designed to be used as a public API.
//   - It's suitable for cases where you need to find all matching elements without waiting.
//   - For more specific element selection (e.g., by index), consider using other methods like ElemByIndex.
//
// See also:
//   - ElemsByText: For text-based element selection.
//   - elems: The internal method used for standard CSS selector-based selection.
func (b *Bot) Elems(selector string, opts ...ElemOptionFunc) ([]*rod.Element, error) {
	// b.mustNotEmpty(selector)
	if selector == "" {
		return nil, ErrSelectorEmpty
	}

	if strings.Contains(selector, SEP) {
		return b.ElemsByText(selector, opts...)
	}

	return b.elems(selector, opts...)
}

// elems retrieves elements matching the given selector, with support for iframe selection and optional timeout.
//
// This internal function is the core implementation for element retrieval, used by higher-level functions
// like Elems. It provides flexible element selection with special handling for iframes and timeout options.
//
// Behavior:
//  1. If the selector contains IFrameSep, it performs iframe-based selection.
//  2. For standard selectors, it uses the page's Elements method.
//  3. When a timeout is specified, it ensures at least one element exists before retrieving all matches.
//
// Parameters:
//   - selector: A string representing the selector to locate the elements.
//   - For standard selection: CSS selector (e.g., "div.class", "#id")
//   - For iframe selection: Uses format "iframeSelector{IFrameSep}contentSelector"
//   - opts: Optional ElemOptionFunc arguments to customize the element search behavior.
//     Notable options include:
//   - WithTimeout: Sets a duration to wait for at least one element to exist.
//
// Returns:
//   - []*rod.Element: A slice of pointers to the found elements. May be empty if no elements are found.
//   - error: An error if elements couldn't be found or if any other issues occurred.
//
// Iframe handling:
//   - If the selector contains IFrameSep, it first locates the iframe, then finds the element within it.
//   - Returns a slice with a single element for iframe-based selection.
//
// Timeout behavior:
//   - If timeout is non-zero, it first calls Elem to ensure at least one element exists.
//   - This can be useful to wait for dynamic content to load before retrieving all matches.
//
// Error handling:
//   - Returns errors from underlying Rod methods (e.g., Element, Elements).
//   - Propagates errors from Elem when using timeout option.
//
// Internal notes:
//   - This function is meant to be used internally by the Bot struct.
//   - It provides the foundational element retrieval logic for other public methods.
//
// Example usage (internal):
//
//	elems, err := b.elems("div.class", WithTimeout(5))
//
// See also:
//   - Elems: The public-facing method that uses this internal function.
//   - IframeElem: Used internally for iframe-based selection.
func (b *Bot) elems(selector string, opts ...ElemOptionFunc) ([]*rod.Element, error) {
	if ss := strings.Split(selector, IFrameSep); len(ss) == _iframeLen {
		elem, err := b.IframeElem(ss[0], ss[1])
		if err != nil {
			return nil, err
		}

		return []*rod.Element{elem}, nil
	}

	opt := ElemOptions{timeout: NapToSec}
	bindElemOptions(&opt, opts...)

	if opt.timeout != 0 {
		// WARN: when timeout not 0, we need to ensure elem existed first,
		// so run GetElem to ensure elem existed or end if err.
		_, err := b.Elem(selector, opts...)
		if err != nil {
			return nil, err
		}
	}

	elems, err := b.page.Elements(selector)
	if err != nil {
		return nil, err
	}

	return elems, nil
}

func (b *Bot) MustAllElemAttrs(selector string, opts ...ElemOptionFunc) []string {
	attrs := []string{}

	elems := b.MustElems(selector, opts...)
	for _, elem := range elems {
		v, err := b.getElementAttr(elem, opts...)
		if err != nil {
			b.logger.Debug("cannot get attr", zap.String("selector", selector), zap.Error(err))
		}

		attrs = append(attrs, v)
	}

	return attrs
}

func (b *Bot) AllAttrs(selector string, opts ...ElemOptionFunc) ([]string, error) {
	attrs := []string{}

	elems, err := b.Elems(selector, opts...)
	if err != nil {
		return nil, err
	}

	for _, elem := range elems {
		v, err := b.getElementAttr(elem, opts...)
		if err != nil {
			return nil, err
		}

		attrs = append(attrs, v)
	}

	return attrs, nil
}

func (b *Bot) MustElemAttr(selector string, opts ...ElemOptionFunc) string {
	v, err := b.ElemAttr(selector, opts...)
	b.pie(err)

	return v
}

// ElemAttr retrieves an attribute value from an element identified by a selector.
//
// By default, retrieves 'innerText'. Use WithAttr(name) to specify a different attribute.
//
// Example:
//
//	href, err := b.ElemAttr("a.link", WithAttr("href"))
//
// Note:
//   - This function is useful for quickly retrieving an attribute value when you
//     only need to perform this single operation on an element.
//   - For multiple operations on the same element, it's more efficient to first
//     select the element using Elem, then perform operations on the returned element.
//
// See also:
//   - Elem: For selecting an element without immediately retrieving an attribute.
//   - MustElemAttr: For a version of this function that panics on error.
//   - ElementAttr: For retrieving an attribute from an already selected element.
func (b *Bot) ElemAttr(selector string, opts ...ElemOptionFunc) (string, error) {
	elem, err := b.Elem(selector, opts...)
	if err != nil {
		return "", err
	}

	return b.getElementAttr(elem, opts...)
}

func (b *Bot) ElementAttr(elem *rod.Element, opts ...ElemOptionFunc) (string, error) {
	return b.getElementAttr(elem, opts...)
}

// getElem retrieves an element using various selection strategies.
//
// This function is a versatile element selector that supports multiple ways to find elements:
//  1. Standard CSS selectors
//  2. Iframe-based selection using IFrameSep
//  3. Text content-based selection using SEP
//  4. Index-based selection for multiple matching elements
//
// The function first checks for an empty selector and returns an error if found.
// It then applies the provided options, with defaults for root and timeout.
//
// Selection process:
//   - If the selector contains IFrameSep, it uses IframeElem to find the element
//   - If the selector contains SEP, it uses ElemByText for text-based selection
//   - If timeout is 0, it uses ElemByIndex for immediate selection without waiting
//   - Otherwise, it waits for the element to appear within the specified timeout
//
// After finding the element, if an index is specified, it retrieves the element at that index.
//
// Parameters:
//   - selector: string to identify the element. Can be a CSS selector, iframe selector, or text content
//   - opts: variadic ElemOptionFunc arguments to customize selection behavior
//
// Returns:
//   - *rod.Element: pointer to the found element
//   - error: nil if successful, otherwise an error describing why the element couldn't be found
//
// The function respects the following options:
//   - WithRoot: to specify a root element for the search
//   - WithTimeout: to set a custom timeout for waiting for the element
//   - WithIndex: to select a specific element when multiple matches are found
//   - WithIframe: to specify an iframe context for the search
//
// Example usage:
//
//	elem, err := b.getElem("div.class", WithTimeout(5), WithIndex(2))
//
// Note: This is an internal function and should be used cautiously outside the Bot struct.
func (b *Bot) getElem(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	// b.mustNotEmpty(selector)
	if selector == "" {
		return nil, ErrSelectorEmpty
	}

	opt := ElemOptions{root: b.root, timeout: ShortToSec}
	bindElemOptions(&opt, opts...)

	if ss := strings.Split(selector, IFrameSep); len(ss) == _iframeLen {
		return b.IframeElem(ss[0], ss[1], opts...)
	}

	// by text content
	if strings.Contains(selector, SEP) {
		return b.ElemByText(selector, opts...)
	}

	// without wait
	if opt.timeout == 0 {
		return b.ElemByIndex(selector, opt.index)
	}

	var (
		elem *rod.Element
		err  error
	)

	dur := time.Duration(opt.timeout) * time.Second

	// this will wait until element shown, or got error
	if opt.root != nil {
		elem, err = opt.root.Timeout(dur).Element(selector)
	} else {
		page := b.page
		if opt.iframe != nil {
			page = opt.iframe
		}

		elem, err = page.Timeout(dur).Element(selector)
	}

	if err != nil {
		return nil, err
	}

	// when index is not 0:
	// this is used when we first need to wait elem to appear, then get the one with index
	if opt.index != 0 {
		elem, err = b.ElemByIndex(selector, opt.index)
	}

	return elem, err
}

func (b *Bot) getElementAttr(elem *rod.Element, opts ...ElemOptionFunc) (string, error) {
	opt := ElemOptions{attr: "innerText"}
	bindElemOptions(&opt, opts...)

	attr := opt.attr
	if attr == "" || attr == "innerText" {
		return elem.Text()
	}

	raw, e := elem.Attribute(attr)
	if e != nil {
		return "", e
	}

	return *raw, nil
}

// AllElementsAttrMap retrieves specified attributes from multiple elements.
//
// This function iterates over a slice of Rod elements and extracts specified
// attributes from each element, returning the results as a slice of maps.
//
// Parameters:
//   - elems: A slice of *rod.Element pointers representing the elements to process.
//   - opts: Optional variadic ElemOptionFunc arguments to customize attribute retrieval.
//     Typically, WithAttrMap is used to specify which attributes to retrieve.
//
// Returns:
//   - []map[string]string: A slice of maps, where each map represents an element's
//     attributes. The keys in the map are the attribute names specified in WithAttrMap,
//     and the values are the corresponding attribute values from the element.
//
// Behavior:
//  1. Iterates through each element in the input slice.
//  2. For each element, calls ElementAttrMap to retrieve the specified attributes.
//  3. Compiles the results into a slice of maps, one map per element.
//
// Options:
//   - WithAttrMap(map[string]string): Specifies which attributes to retrieve and how
//     to name them in the result map. The key is the name in the result, the value
//     is the actual attribute name to retrieve from the element.
//
// Usage example:
//
//	elements, _ := b.Elems("a.external-link")
//	attrMap := map[string]string{
//	    "text": "innerText",
//	    "link": "href",
//	    "title": "title",
//	}
//	results := b.AllElementsAttrMap(elements, WithAttrMap(attrMap))
//	for _, elem := range results {
//	    fmt.Printf("Link: %s, Text: %s, Title: %s\n", elem["link"], elem["text"], elem["title"])
//	}
//
// Note:
//   - This function is useful for bulk attribute retrieval from multiple elements,
//     especially when you need to collect the same set of attributes from each element.
//   - If an attribute specified in WithAttrMap is not present on an element, its value
//     in the result map will be an empty string.
//   - Any errors encountered while retrieving attributes from individual elements are
//     logged as warnings but do not stop the process for other elements.
//
// See also:
//   - ElementAttrMap: For retrieving attributes from a single element.
//   - Elems: For selecting multiple elements to use with this function.
func (b *Bot) AllElementsAttrMap(elems []*rod.Element, opts ...ElemOptionFunc) []map[string]string {
	var output []map[string]string
	for _, elem := range elems {
		output = append(output, b.ElementAttrMap(elem, opts...))
	}

	return output
}

// ElementAttrMap retrieves specified attributes from a single element.
//
// This function extracts multiple attributes from a given Rod element and
// returns them as a map of key-value pairs.
//
// Behavior:
//  1. Initializes an empty result map.
//  2. Applies the provided options to determine which attributes to retrieve.
//  3. For each specified attribute, calls getElementAttr to retrieve its value.
//  4. Compiles the results into a map.
//
// Options:
//   - WithAttrMap(map[string]string): Specifies which attributes to retrieve and how
//     to name them in the result map. The key is the name in the result, the value
//     is the actual attribute name to retrieve from the element.
//
// Usage example:
//
//	element, _ := b.Elem("a.external-link")
//	attrMap := map[string]string{
//	    "text": "innerText",
//	    "link": "href",
//	    "title": "title",
//	}
//	result := b.ElementAttrMap(element, WithAttrMap(attrMap))
//	fmt.Printf("Link: %s, Text: %s, Title: %s\n", result["link"], result["text"], result["title"])
//
// Error handling:
//   - If an attribute cannot be retrieved, a warning is logged, and the attribute
//     is omitted from the result map.
//   - The function does not return an error; it attempts to retrieve as many
//     attributes as possible and logs warnings for any that fail.
//
// Note:
//   - This function is useful when you need to retrieve multiple attributes
//     from a single element in one operation.
//   - If an attribute specified in WithAttrMap is not present on the element,
//     it will be omitted from the result map and a warning will be logged.
//   - The 'innerText' attribute is handled specially and retrieved using the Text() method.
//
// See also:
//   - getElementAttr: The underlying method used to retrieve individual attributes.
//   - AllElementsAttrMap: For retrieving attributes from multiple elements at once.
func (b *Bot) ElementAttrMap(elem *rod.Element, opts ...ElemOptionFunc) map[string]string {
	opt := ElemOptions{}
	bindElemOptions(&opt, opts...)

	res := make(map[string]string)

	for key, attr := range opt.attrMap {
		raw, err := b.getElementAttr(elem, WithAttr(attr))
		if err != nil {
			b.logger.Sugar().Warnf("cannot get attr (%s:%s), %v", key, attr, zap.Error(err))
		}

		res[key] = raw
	}

	return res
}

func (b *Bot) mustNotEmpty(selector string) { //nolint:unused
	if selector == "" {
		const callerStackOffset = 2
		w, i := xpretty.Caller(callerStackOffset)
		b.logger.Sugar().Fatalf("selector is empty: file=%s,line=%d", w, i)
	}
}

func (b *Bot) mustNotByText(selector string) { //nolint:unused
	if strings.Contains(selector, SEP) {
		const callerStackOffset = 2
		w, i := xpretty.Caller(callerStackOffset)
		b.logger.Sugar().Fatalf("never use selector by text: file=%s,line=%d", w, i)
	}
}

func (b *Bot) MustAnyElem(selectors []string, opts ...ElemOptionFunc) string {
	start := time.Now()
	sel, err := b.AnyElem(selectors, opts...)
	b.pie(err)

	cost := time.Since(start).Seconds()
	if cost > _logIfTimeout {
		b.logger.Sugar().Debugf("get selector:(%s) by ensure cost=%.2fs", sel, cost)
	}

	return sel
}

// AnyElem attempts to find any element from a list of selectors within a specified timeout.
//
// This function is useful when you're looking for one of several possible elements
// on a page, and you don't know which one will appear first or at all.
//
// Parameters:
//   - selectors: A slice of strings, each representing a CSS selector or a text-based
//     selector (using SEP format) to search for.
//   - opts: Optional variadic ElemOptionFunc arguments to customize the search behavior.
//
// Returns:
//   - string: The selector of the first element found.
//   - error: An error if no element could be found within the timeout period or if
//     other issues occurred during the search.
//
// Behavior:
//  1. Sets up default options (timeout: MediumToSec, retries: 1) which can be overridden.
//  2. Uses Rod's Race method to concurrently search for all provided selectors.
//  3. Returns as soon as any of the selectors matches an element.
//  4. If no element is found within the timeout, it retries based on the retry option.
//
// Options:
//   - WithTimeout(duration): Sets the maximum time to wait for any element to appear.
//   - WithRetries(count): Sets the number of times to retry the entire search if no
//     element is found.
//   - Other standard ElemOptions can also be used and will be applied to each selector search.
//
// Usage example:
//
//	selectors := []string{"#loading", ".content", "button@@@Submit"}
//	foundSelector, err := b.AnyElem(selectors, WithTimeout(10), WithRetries(3))
//	if err != nil {
//	    // Handle error: no element found
//	} else {
//	    fmt.Printf("Found element with selector: %s\n", foundSelector)
//	}
//
// Note:
//   - This function is particularly useful for handling dynamic content where
//     different elements might appear depending on the state of the page.
//   - It supports both standard CSS selectors and text-based selectors (using SEP format).
//   - The function logs a debug message if finding the element takes longer than _logIfTimeout.
//
// Error handling:
//   - Returns an error if no element is found after all retries are exhausted.
//   - Any errors encountered during the search process are propagated back to the caller.
//
// See also:
//   - MustAnyElem: A version of this function that panics on error.
//   - Elem: For finding a single specific element.
//   - AnyElemAttribute: If you need to retrieve an attribute from the found element.
func (b *Bot) AnyElem(selectors []string, opts ...ElemOptionFunc) (string, error) {
	opt := ElemOptions{timeout: MediumToSec, retries: 1}
	bindElemOptions(&opt, opts...)

	var (
		sel string
		err error
	)

	err = retry.Do(
		func() error {
			// err = rod.Try(func() {
			race := b.page.Timeout(time.Duration(opt.timeout) * time.Second).Race()
			for _, s := range selectors {
				b.appendToRace(s, &sel, race)
			}

			_, err = race.Do()

			return err
		},
		retry.Attempts(opt.retries),
		retry.LastErrorOnly(true),
	)

	return sel, err
}

// appendToRace adds a selector to a Rod race context, handling both standard CSS selectors
// and text-based selectors.
//
// This function is an internal helper used by AnyElem to set up concurrent element searches.
// It differentiates between standard CSS selectors and text-based selectors (containing SEP),
// and sets up the appropriate race condition for each.
//
// Parameters:
//   - selector: A string representing either a CSS selector or a text-based selector.
//     Text-based selectors should be in the format "cssSelector@@@textContent".
//   - out: A pointer to a string where the matching selector will be stored when found.
//   - race: A pointer to a Rod RaceContext to which the search condition will be added.
//
// Behavior:
//  1. Checks if the selector contains the SEP ("@@@") to determine if it's a text-based selector.
//  2. For text-based selectors:
//     - Splits the selector into CSS and text components.
//     - Uses Rod's ElementR method to search for an element matching both CSS and text content.
//  3. For standard CSS selectors:
//     - Uses Rod's Element method to search for the element.
//  4. In both cases, when an element is found, the original selector is stored in the 'out' parameter.
//
// Usage:
//
//	This function is typically not called directly but is used internally by AnyElem.
//	It's designed to be used with Rod's Race feature for concurrent element searches.
//
// Example (internal usage):
//
//	race := b.page.Timeout(duration).Race()
//	var foundSelector string
//	for _, selector := range selectors {
//	    b.appendToRace(selector, &foundSelector, race)
//	}
//	_, err := race.Do()
//
// Note:
//   - This function is crucial for enabling AnyElem to work with both CSS and text-based selectors.
//   - It encapsulates the logic for differentiating between selector types and setting up
//     the appropriate race condition for each.
//   - The function doesn't return anything; it modifies the race context and sets up a callback
//     to store the found selector.
//
// See also:
//   - AnyElem: The main function that uses appendToRace to set up concurrent element searches.
//   - Rod RaceContext: The underlying Rod feature used for concurrent operations.
func (b *Bot) appendToRace(selector string, out *string, race *rod.RaceContext) {
	if strings.Contains(selector, SEP) {
		ss := strings.Split(selector, SEP)
		txt := strings.Join(ss[1:], SEP)
		race.ElementR(ss[0], txt).MustHandle(func(_ *rod.Element) {
			*out = selector
		})

		return
	}

	race.Element(selector).MustHandle(func(_ *rod.Element) {
		*out = selector
	})
}

func (b *Bot) MustIframeElem(iframe, selector string, opts ...ElemOptionFunc) *rod.Element {
	elem, err := b.IframeElem(iframe, selector, opts...)
	b.pie(err)

	return elem
}

// IframeElem finds and returns an element within an iframe.
//
// This function performs a two-step element selection:
// 1. It locates the iframe element in the main document.
// 2. It then searches for the specified element within the iframe's content.
//
// Parameters:
//   - iframe: A string selector to locate the iframe element in the main document.
//     This should be a valid CSS selector targeting the iframe.
//   - selector: A string CSS selector to find the desired element within the iframe's content.
//   - opts: Optional variadic ElemOptionFunc arguments to customize the element search behavior.
//     These options are applied to both the iframe selection and the inner element selection.
//
// Returns:
//   - *rod.Element: A pointer to the found element within the iframe.
//   - error: An error if either the iframe or the inner element couldn't be found,
//     or if any other issues occurred during the process.
//
// The function uses the Bot's Elem method to find the iframe, ensuring all standard
// element selection options and timeouts are respected. It then creates a new Rod Frame
// from the iframe element and searches for the inner element within this frame.
//
// Usage example:
//
//	elem, err := b.IframeElem("#my-iframe", ".content-class")
//	if err != nil {
//	    // Handle error
//	}
//	// Use elem
//
// Note: This function is useful for interacting with elements that are within iframes,
// which are not directly accessible from the main document context.
func (b *Bot) IframeElem(iframe, selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	iframeElem, err := b.Elem(iframe, opts...)
	if err != nil {
		return nil, err
	}

	opts = append(opts, WithIframe(iframeElem.MustFrame()))

	return b.Elem(selector, opts...)
}

// EnsureNonNilElem retrieves a non-nil element by selector, ensuring its existence and validity.
//
// This function combines element retrieval with a nil check, providing a robust way
// to fetch an element and verify its presence in the DOM. It's particularly useful
// for operations that require a guaranteed non-nil element, such as form interactions
// or complex UI manipulations.
//
// Returns:
//   - *rod.Element: A pointer to the found Rod element, guaranteed to be non-nil if no error is returned.
//   - error: An error is returned in the following cases:
//     1. If the underlying Elem method fails to find the element (e.g., timeout, invalid selector).
//     2. If the element is found but is nil (which could happen in certain edge cases with dynamic content).
//
// Behavior:
//  1. Calls the underlying Elem method with the provided selector and options.
//  2. If Elem returns an error, that error is immediately propagated back to the caller.
//  3. If Elem succeeds but returns a nil element, an appropriate error (likely ErrCannotFindSelector) is returned.
//  4. If a non-nil element is found, it is returned along with a nil error.
//
// Usage example:
//
//	elem, err := b.EnsureNonNilElem("#login-button", []ElemOptionFunc{WithTimeout(10)})
//	if err != nil {
//	    if errors.Is(err, ErrCannotFindSelector) {
//	        log.Println("Login button not found or is nil")
//	    } else {
//	        log.Println("Error finding login button:", err)
//	    }
//	    return
//	}
//	// Safely interact with the non-nil element
//	err = elem.Click()
//
// Error handling:
//   - Propagates any errors from the underlying Elem method, which could include
//     timeout errors, selector syntax errors, or other Rod-specific errors.
//   - Returns a specific error (likely ErrCannotFindSelector) if the element is found but nil,
//     allowing for differentiated handling of "not found" vs "found but nil" scenarios.
//
// Best practices:
//   - Use this function when you need to ensure an element's presence before performing critical operations.
//   - Combine with appropriate timeout options to balance between waiting for dynamic content and failing fast.
//   - Handle both the general error case and the specific ErrCannotFindSelector case in your error handling logic.
//
// Note:
//   - This function is essential for robust web automation scripts, as it helps prevent
//     nil pointer dereference errors that could occur if an element is unexpectedly absent.
//   - It's particularly valuable in scenarios with dynamic content where elements might
//     not be immediately available or could be temporarily removed from the DOM.
//
// See also:
//   - Elem: The underlying method used for basic element retrieval.
//   - ElemOptions: For understanding the available options for element selection.
//   - ErrCannotFindSelector: The likely error type returned for nil elements.
func (b *Bot) EnsureNonNilElem(sel string, opts []ElemOptionFunc) (*rod.Element, error) {
	elem, err := b.Elem(sel, opts...)
	if err != nil {
		return nil, err
	}

	if elem == nil {
		return nil, ErrCannotFindSelector(sel)
	}

	return elem, nil
}

// TryFocusElement attempts to focus on a given element, with optional scrolling to the right.
//
// This function tries to make an element interactable by focusing on it and, if necessary,
// scrolling the page. It's particularly useful for elements that might be out of view
// or require scrolling to become accessible.
//
// Behavior:
//  1. Attempts to check if the element is interactable.
//  2. If not interactable, it tries to scroll to the element.
//  3. If scrollToRight is true, it also scrolls horizontally.
//  4. This process is retried up to 3 times (using RetryIn3).
//
// The function performs the following steps in each attempt:
//  1. Checks if the element is interactable using elem.Interactable().
//  2. If not interactable:
//     - Scrolls to the element using b.ScrollToElemDirectly(elem).
//     - If scrollToRight is true, scrolls 1024 pixels to the right.
//  3. If the element becomes interactable, the function returns nil.
//  4. If it remains non-interactable after 3 attempts, it returns an error.
//
// Usage example:
//
//	elem, _ := b.Elem("#my-button")
//	err := b.TryFocusElement(elem, false)
//	if err != nil {
//	    log.Println("Failed to focus on element:", err)
//	    return
//	}
//	// Element is now focused and interactable
//	elem.Click()
//
// Error handling:
//   - Returns ErrNotInteractable if the element cannot be made interactable after 3 attempts.
//   - Any errors from scrolling operations are ignored within the retry loop.
//
// Note:
//   - This function is useful for handling elements that may be initially out of view
//     or require page manipulation to become accessible.
//   - The scrolling behavior uses direct page manipulation, which might not trigger
//     certain JavaScript events associated with normal scrolling.
//   - The horizontal scroll amount (1024 pixels) is hardcoded and may need adjustment
//     for specific use cases.
//
// See also:
//   - ScrollToElemDirectly: The method used for scrolling to the element.
//   - RetryIn3: The retry mechanism used in this function.
func (b *Bot) TryFocusElement(elem *rod.Element, scrollToRight bool) error {
	return RetryIn3(
		func() error {
			if _, err := elem.Interactable(); err != nil {
				_ = rod.Try(func() {
					_ = b.ScrollToElemDirectly(elem)
					if scrollToRight {
						_ = b.page.Mouse.Scroll(1024, 0, 4) //nolint:mnd
					}
				})

				return err
			}

			return nil
		})
}

// OpenOrClickElement attempts to open or interact with a given element, with options for opening in a new tab.
//
// This function provides a flexible way to interact with elements, particularly links or
// buttons that lead to new content. It can either click the element directly or open it
// in a new tab, depending on the options provided.
//
// Parameters:
//   - elem: A *rod.Element pointer representing the element to open or interact with.
//   - opts: Variadic ElemOptionFunc arguments to customize the opening behavior.
//     Notable options include:
//   - WithOpenInTab(bool): If true, opens the element in a new tab instead of clicking it directly.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
//
// Behavior:
//  1. Applies the provided options to determine the interaction method.
//  2. Attempts to focus on the element using TryFocusElem.
//  3. If focusing fails, returns ErrNotInteractable.
//  4. If WithOpenInTab is true, calls OpenInNewTab.
//  5. Otherwise, calls ClickElem for a standard click interaction.
//
// Usage example:
//
//	elem, _ := b.Elem("a.external-link")
//	err := b.OpenOrClickElement(elem, WithOpenInTab(true))
//	if err != nil {
//	    if errors.Is(err, ErrNotInteractable) {
//	        log.Println("Element is not interactable")
//	    } else {
//	        log.Println("Failed to open element:", err)
//	    }
//	    return
//	}
//	// Element has been opened in a new tab
//
// Error handling:
//   - Returns ErrNotInteractable if the element cannot be focused or made interactable.
//   - Propagates errors from OpenInNewTab or ClickElem operations.
//
// Options:
//   - WithOpenInTab(bool): Determines whether to open the element in a new tab.
//   - Other standard ElemOptions can be used to customize the element interaction.
//
// Note:
//   - This function is particularly useful for handling navigation elements or
//     interactive elements that may open new content.
//   - The behavior of opening in a new tab may depend on the browser's settings
//     and the website's implementation.
//   - Ensure that the browser instance supports multiple tabs when using WithOpenInTab.
//
// See also:
//   - TryFocusElem: Used internally to ensure the element is interactable.
//   - OpenInNewTab: Called when opening the element in a new tab.
//   - ClickElem: Used for standard click interactions.
func (b *Bot) OpenOrClickElement(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{root: b.root, timeout: ShortToSec}
	bindElemOptions(&opt, opts...)

	if err := b.TryFocusElement(elem, true); err != nil {
		return ErrNotInteractable
	}

	if opt.openInTab {
		return b.OpenElementInNewTab(elem)
	}

	return b.ClickElem(elem)
}

// OpenElementInNewTab opens an element (usually a link or button) in a new browser tab.
//
// This function simulates a Ctrl+Click (or Cmd+Click on macOS) on the given element,
// which typically opens the link in a new tab in most web browsers. After opening
// the new tab, it switches the bot's focus to the newly opened tab.
//
// Parameters:
//   - elem: A *rod.Element pointer representing the element to be opened in a new tab.
//     This is typically a link (<a> tag) or a button that triggers navigation.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
//
// Behavior:
//  1. Retrieves the current list of open pages in the browser.
//  2. Determines the appropriate key for Ctrl/Cmd based on the operating system.
//  3. Focuses on and highlights the target element.
//  4. Simulates a Ctrl/Cmd + Enter key press on the element.
//  5. Waits for and activates the newly opened page.
//
// Usage example:
//
//	linkElem, _ := b.Elem("a.external-link")
//	err := b.OpenElementInNewTab(linkElem)
//	if err != nil {
//	    log.Println("Failed to open link in new tab:", err)
//	    return
//	}
//	// Bot is now focused on the newly opened tab
//
// Error handling:
//   - Returns an error if the Ctrl/Cmd + Enter key action fails.
//   - Returns an error if unable to switch to the newly opened page.
//
// Note:
//   - This function assumes that Ctrl/Cmd + Enter will open the link in a new tab,
//     which is standard behavior for most websites but may not work in all cases.
//   - The function waits for up to 10 seconds for the new tab to open before timing out.
//   - Ensure that pop-ups are not blocked in the browser settings, as this might
//     prevent new tabs from opening.
//   - This function automatically switches the bot's context to the new tab.
//
// OS Compatibility:
//   - Automatically uses Cmd key on macOS and Ctrl key on other operating systems.
//
// See also:
//   - FocusAndHighlight: Used internally to focus on the element before interaction.
//   - ActivateLastOpenedPage: Used to switch to the newly opened page.
func (b *Bot) OpenElementInNewTab(elem *rod.Element) error {
	pages := b.browser.MustPages()

	ctrlKey := input.ControlLeft
	if sysutil.IsMac() {
		ctrlKey = input.MetaLeft
	}

	b.FocusAndHighlight(elem)

	err := elem.MustKeyActions().Press(ctrlKey).Type(input.Enter).Do()
	if err != nil {
		return fmt.Errorf("cannot do ctrl click: %w", err)
	}

	err = b.ActivateLastOpenedPage(pages, 10) //nolint:mnd
	if err != nil {
		return fmt.Errorf("cannot switch opened page: %w", err)
	}

	return nil
}
