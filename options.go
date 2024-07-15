package wee

import "github.com/go-rod/rod"

type ScrollAsHuman struct {
	enabled          bool
	longSleepChance  float64
	shortSleepChance float64
	scrollUpChance   float64
}

type ElemOptions struct {
	root   *rod.Element
	iframe *rod.Page

	caseInsensitive bool

	// elem
	index   int
	attr    string
	attrMap map[string]string

	selectorType rod.SelectorType

	timeout float64
	retries uint

	submit    bool
	humanized bool

	trigger          bool
	clearBeforeInput bool
	endWithEscape    bool

	waitStable bool

	clickByScript    bool
	handleCoverByEsc bool

	scrollAsHuman *ScrollAsHuman

	highlight bool

	// scroll setup
	steps       int
	offsetToTop float64
}

type ElemOptionFunc func(o *ElemOptions)

func bindElemOptions(opt *ElemOptions, opts ...ElemOptionFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithRoot(root *rod.Element) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.root = root
	}
}

func WithIframe(page *rod.Page) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.iframe = page
	}
}

func WithCaseInsensitive(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.caseInsensitive = b
	}
}

func WithTimeout(t float64) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.timeout = t
	}
}

func WithRetries(i uint) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.retries = i
	}
}

func WithIndex(i int) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.index = i
	}
}

func WithAttr(s string) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.attr = s
	}
}

func WithAttrMap(m map[string]string) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.attrMap = m
	}
}

func WithSelectorType(t rod.SelectorType) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.selectorType = t
	}
}

func WithSubmit(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.submit = b
	}
}

func Trigger(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.trigger = b
	}
}

func ClearBeforeInput(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.clearBeforeInput = b
	}
}

func EndWithEscape(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.endWithEscape = b
	}
}

func WithHumanized(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.humanized = b
	}
}

func WithWaitStable(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.waitStable = b
	}
}

func WithClickByScript(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.clickByScript = b
	}
}

func WithHighlight(b bool) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.highlight = b
	}
}

func DisableHandleCoverByEsc() ElemOptionFunc {
	return func(o *ElemOptions) {
		o.handleCoverByEsc = false
	}
}

func WithScrollAsHuman(ah *ScrollAsHuman) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.scrollAsHuman = ah
	}
}

func WithSteps(i int) ElemOptionFunc {
	return func(o *ElemOptions) {
		o.steps = i
	}
}
