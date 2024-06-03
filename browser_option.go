package wee

import "github.com/samber/lo"

type BrowserOptions struct {
	slowDelay  int
	paintRects bool
	headless   bool
	flags      []string
}

type BrowserOptionFunc func(o *BrowserOptions)

func bindBrowserOptions(opt *BrowserOptions, opts ...BrowserOptionFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithPaintRects(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.paintRects = b
	}
}

func WithSlowDelay(i int) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.slowDelay = i
	}
}

func WithFlags(arr ...string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.flags = append(o.flags, arr...)
		o.flags = lo.Uniq(o.flags)
	}
}

func WithBrowserHeadless(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.headless = b
	}
}
