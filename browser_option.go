package wee

import "github.com/samber/lo"

type BrowserOptions struct {
	slowDelay  int
	paintRects bool
	headless   bool
	flags      []string

	extensions []string

	noDefaultDevice bool
	incognito       bool
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

func NoDefaultDevice(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.noDefaultDevice = b
	}
}

func Incognito(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.incognito = b
	}
}

func WithExtensions(extArr []string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.extensions = extArr
	}
}
