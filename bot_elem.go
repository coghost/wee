package wee

import (
	"fmt"
	"strings"
	"time"

	"github.com/coghost/xpretty"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

const (
	_logIfTimeout = 5.0
	_iframeLen    = 2
)

func (b *Bot) MustGetElem(selector string, opts ...ElemOptionFunc) *rod.Element {
	elem, err := b.GetElem(selector, opts...)
	b.e(err)

	return elem
}

// GetElem get element by selector.
func (b *Bot) GetElem(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	start := time.Now()

	elem, err := b.getElem(selector, opts...)
	if err != nil {
		return nil, err
	}

	if cost := time.Since(start).Seconds(); cost > _logIfTimeout {
		log.Debug().Float64("cost", cost).Str("selector", selector).Msg("GetElem")
	}

	return elem, nil
}

// GetElemInstantly get elem without wait.
func (b *Bot) GetElemInstantly(selector string, index int) (*rod.Element, error) {
	elems, err := b.GetElems(selector)
	if err != nil {
		return nil, err
	}

	index = refineIndex(len(elems), index)
	if index < 0 {
		return nil, nil
	}

	return elems[index], nil
}

// GetElemByText by text content.
//
// Available layouts:
//   - when selector is like `div.abc@@@txt`, will use contains
//   - when selector is like `div.abc@@@---@@@txt`, will use exact match
func (b *Bot) GetElemByText(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	opt := ElemOptions{root: b.root}
	bindElemOptions(&opt, opts...)

	arr := strings.Split(selector, SEP)
	txt := arr[len(arr)-1]

	const exactMatchLen = 3

	if len(arr) == exactMatchLen {
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

func (b *Bot) MustGetElemsOfAllSelectors(selectors []string, opts ...ElemOptionFunc) []*rod.Element {
	var elems []*rod.Element
	for _, s := range selectors {
		elems = append(elems, b.MustGetElems(s)...)
	}

	return elems
}

func (b *Bot) MustGetElems(selector string, opts ...ElemOptionFunc) []*rod.Element {
	elems, err := b.GetElems(selector, opts...)
	b.e(err)

	return elems
}

// GetElems get all elements immediately.
func (b *Bot) GetElems(selector string, opts ...ElemOptionFunc) ([]*rod.Element, error) {
	b.mustNotEmpty(selector)
	b.mustNotByText(selector)

	if ss := strings.Split(selector, IFrameSep); len(ss) == _iframeLen {
		elem, err := b.GetIframeElem(ss[0], ss[1])
		if err != nil {
			return nil, err
		}

		return []*rod.Element{elem}, nil
	}

	opt := ElemOptions{timeout: 0}
	bindElemOptions(&opt, opts...)

	if opt.timeout != 0 {
		// WARN: when timeout not 0, we need to ensure elem existed first,
		// so run GetElem to ensure elem existed or end if err.
		_, err := b.GetElem(selector, opts...)
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

// GetElemsWithShortTo set the timeout to `5 seconds`
func (b *Bot) GetElemsWithShortTo(selector string, opts ...ElemOptionFunc) ([]*rod.Element, error) {
	opts = append(opts, WithTimeout(ShortToSec))

	return b.GetElems(selector, opts...)
}

func (b *Bot) MustGetAllElemsAttrs(selector string, opts ...ElemOptionFunc) []string {
	attrs := []string{}

	elems := b.MustGetElems(selector, opts...)
	for _, elem := range elems {
		v, err := b.getElementAttr(elem, opts...)
		if err != nil {
			log.Debug().Err(err).Msg("cannot get attr")
		}

		attrs = append(attrs, v)
	}

	return attrs
}

func (b *Bot) MustGetElemAttr(selector string, opts ...ElemOptionFunc) string {
	v, err := b.GetElemAttr(selector, opts...)
	b.e(err)

	return v
}

func (b *Bot) GetElemAttr(selector string, opts ...ElemOptionFunc) (string, error) {
	elem, err := b.GetElem(selector, opts...)
	if err != nil {
		return "", err
	}

	return b.getElementAttr(elem, opts...)
}

func (b *Bot) GetElementAttr(elem *rod.Element, opts ...ElemOptionFunc) (string, error) {
	return b.getElementAttr(elem, opts...)
}

func (b *Bot) getElem(selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	b.mustNotEmpty(selector)

	opt := ElemOptions{root: b.root, timeout: MediumToSec}
	bindElemOptions(&opt, opts...)

	if ss := strings.Split(selector, IFrameSep); len(ss) == _iframeLen {
		return b.GetIframeElem(ss[0], ss[1], opts...)
	}

	// by text content
	if strings.Contains(selector, SEP) {
		return b.GetElemByText(selector, opts...)
	}

	// without wait
	if opt.timeout == 0 {
		return b.GetElemInstantly(selector, opt.index)
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
		elem, err = b.GetElemInstantly(selector, opt.index)
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

	s, e := elem.Attribute(attr)

	if e != nil {
		return "", e
	}

	return *s, nil
}

func (b *Bot) GetAllElemsAttrMap(elems []*rod.Element, opts ...ElemOptionFunc) []map[string]string {
	var output []map[string]string
	for _, elem := range elems {
		output = append(output, b.GetElementAttrMap(elem, opts...))
	}

	return output
}

func (b *Bot) GetElementAttrMap(elem *rod.Element, opts ...ElemOptionFunc) map[string]string {
	opt := ElemOptions{}
	bindElemOptions(&opt, opts...)

	res := make(map[string]string)

	for k, attr := range opt.attrMap {
		v, err := b.getElementAttr(elem, WithAttr(attr))
		if err != nil {
			log.Error().Err(err).Msg("cannot get attr for attr map")
		}

		res[k] = v
	}

	return res
}

func (b *Bot) mustNotEmpty(selector string) {
	if selector == "" {
		const callerStackOffset = 2
		w, i := xpretty.Caller(callerStackOffset)
		log.Fatal().Str("file", w).Int("line", i).Msg("selector is empty")
	}
}

func (b *Bot) mustNotByText(selector string) {
	if strings.Contains(selector, SEP) {
		const callerStackOffset = 2
		w, i := xpretty.Caller(callerStackOffset)
		log.Fatal().Str("file", w).Int("line", i).Msg("cannot use selector by text")
	}
}

func (b *Bot) MustEnsureAnyElem(selectors ...string) string {
	start := time.Now()
	s, err := b.EnsureAnyElem(selectors...)
	b.e(err)

	cost := time.Since(start).Seconds()
	if cost > _logIfTimeout {
		log.Debug().Str("selector", s).Str("cost", fmt.Sprintf("%.3fs", cost)).Msg("get by ensure")
	}

	return s
}

func (b *Bot) EnsureAnyElem(selectors ...string) (string, error) {
	return b.EnsureAnyElemWithTimeout(selectors)
}

func (b *Bot) EnsureAnyElemWithTimeout(selectors []string, opts ...ElemOptionFunc) (string, error) {
	opt := ElemOptions{timeout: MediumToSec}
	bindElemOptions(&opt, opts...)

	var (
		sel string
		err error
	)

	err = rod.Try(func() {
		r := b.page.Timeout(time.Duration(opt.timeout) * time.Second).Race()
		for _, s := range selectors {
			b.appendToRace(s, &sel, r)
		}
		r.MustDo()
	})

	return sel, err
}

// appendToRace:
// if directly add race.Element in EnsureAnyElem, will always return the
// last of the selectors
func (b *Bot) appendToRace(selector string, out *string, race *rod.RaceContext) {
	if strings.Contains(selector, SEP) {
		ss := strings.Split(selector, SEP)
		txt := strings.Join(ss[1:], SEP)
		race.ElementR(ss[0], txt).MustHandle(func(_ *rod.Element) {
			*out = selector
		})
	} else {
		race.Element(selector).MustHandle(func(_ *rod.Element) {
			*out = selector
		})
	}
}

func (b *Bot) MustGetIframeElem(iframe, selector string, opts ...ElemOptionFunc) *rod.Element {
	elem, err := b.GetIframeElem(iframe, selector, opts...)
	b.e(err)

	return elem
}

func (b *Bot) GetIframeElem(iframe, selector string, opts ...ElemOptionFunc) (*rod.Element, error) {
	iframeElem, err := b.GetElem(iframe, opts...)
	if err != nil {
		return nil, err
	}

	opts = append(opts, WithIframe(iframeElem.MustFrame()))

	return b.GetElem(selector, opts...)
}
