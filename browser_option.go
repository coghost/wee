package wee

import "github.com/samber/lo"

type BrowserOptions struct {
	paintRects      bool
	headless        bool
	noDefaultDevice bool
	incognito       bool

	ignoreCertErrors bool

	slowMotionDelay int

	userDataDir string
	// proxy ip:port
	proxy string

	flags []string
	// extensions dirs for unpacked extension
	extensions []string
}

type BrowserOptionFunc func(o *BrowserOptions)

func bindBrowserOptions(opt *BrowserOptions, opts ...BrowserOptionFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func BrowserPaintRects(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.paintRects = b
	}
}

func BrowserIgnoreCertErrors(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.ignoreCertErrors = b
	}
}

func BrowserSlowMotionDelay(i int) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.slowMotionDelay = i
	}
}

// BrowserFlags flags those are not pre-defined
func BrowserFlags(arr ...string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.flags = append(o.flags, arr...)
		o.flags = lo.Uniq(o.flags)
	}
}

func BrowserUserDataDir(s string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.userDataDir = s
	}
}

func BrowserHeadless(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.headless = b
	}
}

func BrowserNoDefaultDevice(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.noDefaultDevice = b
	}
}

func BrowserProxy(s string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.proxy = s
	}
}

func BrowserIncognito(b bool) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.incognito = b
	}
}

// BrowserExtensions dirs of extensions, only unpacked extension is supported.
func BrowserExtensions(extArr ...string) BrowserOptionFunc {
	return func(o *BrowserOptions) {
		o.extensions = extArr
	}
}
