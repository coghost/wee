package wee

import (
	"context"
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
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
	b.pie(b.Click(selector, opts...))
}

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

func (b *Bot) ClickElem(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{handleCoverByEsc: true, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.clickByScript {
		return b.ClickElemWithScript(elem, opts...)
	}

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	err := b.EnsureInteractable(elem, opt.handleCoverByEsc)
	if err != nil {
		return err
	}

	err = elem.Timeout(b.shortToSec).Click(proto.InputMouseButtonLeft, 1)
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, &rod.InvisibleShapeError{}) {
		return b.ClickElemWithScript(elem, opts...)
	}

	return err
}

func (b *Bot) EnsureInteractable(elem *rod.Element, byEsc bool) error {
	err := b.Interactable(elem)
	if err == nil {
		return nil
	}

	if hit := b.CloseIfHasPopovers(b.popovers...); hit != 0 {
		log.Trace().Int("hit", hit).Msg("closed pop")
		return nil
	}

	if errors.Is(err, &rod.CoveredError{}) && byEsc {
		return b.PressEscape(elem)
	}

	return err
}

func (b *Bot) Interactable(elem *rod.Element) error {
	if _, err := elem.Interactable(); err != nil {
		e := elem.ScrollIntoView()
		if e == nil {
			if _, err = elem.Interactable(); err != nil {
				return err
			}
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
	return elem.Timeout(b.shortToSec).MustKeyActions().Press(keys...).Do()
}

func (b *Bot) ClickWithScript(selector string, opts ...ElemOptionFunc) error {
	elem, err := b.Elem(selector, opts...)
	if err != nil {
		return err
	}

	return b.ClickElemWithScript(elem, opts...)
}

func (b *Bot) ClickElemWithScript(elem *rod.Element, opts ...ElemOptionFunc) error {
	log.Trace().Msg("start click elem with script")
	opt := ElemOptions{timeout: MediumToSec, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	_, err := elem.CancelTimeout().Timeout(time.Duration(opt.timeout)*time.Second).Eval(`(elem) => { this.click() }`, elem)
	if err != nil {
		log.Error().Err(err).Msg("Err: close by Eval script this.click()")
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
	if len(cookies) == 0 {
		return
	}

	for _, sel := range cookies {
		b.MustClick(sel)
	}

	b.MustWaitLoad()
}

func (b *Bot) CloseIfHasPopovers(popovers ...string) int {
	hit := 0

	if len(popovers) == 0 {
		return 0
	}

	for _, pop := range popovers {
		n, err := b.ClosePopover(pop)
		if err == nil {
			hit += n
		} else {
			log.Debug().Err(err).Msg("cannot close pop")
		}
	}

	if hit != 0 {
		log.Debug().Int("count", hit).Msg("closed popovers")
	}

	return hit
}

func (b *Bot) ClosePopover(sel string) (int, error) {
	hit := 0

	elems, err := b.Elems(sel)
	if err != nil {
		return 0, err
	}

	if len(elems) == 0 {
		log.Trace().Msg("no popovers found")
		return 0, nil
	}

	for _, elem := range elems {
		log.Trace().Str("popover", sel).Msg("try close")
		if _, err := elem.Interactable(); err != nil {
			// elem.Overlay("popover is not interactable")
			return 0, err
		}

		b.HighlightElem(elem)

		e := elem.Click(proto.InputMouseButtonLeft, 1)
		if e != nil {
			return 0, e
		}

		hit += 1
	}

	return hit, nil
}

func (b *Bot) MustClickOneByOne(selectors ...string) {
	for _, sel := range selectors {
		b.MustClick(sel)
		RandSleep(0.5, 0.6)
	}

	b.MustStable()
}
