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

func (b *Bot) MustClickElemAndWait(selector string, opts ...ElemOptionFunc) {
	b.MustClickElem(selector, opts...)
	b.page.MustWaitStable()
}

func (b *Bot) MustClickElem(selector string, opts ...ElemOptionFunc) {
	b.e(b.ClickElem(selector, opts...))
}

func (b *Bot) ClickElem(selector string, opts ...ElemOptionFunc) error {
	elem, err := b.GetElem(selector, opts...)
	if err != nil {
		return err
	}

	if elem == nil {
		return ErrCannotFindElem(selector)
	}

	return b.ClickElement(elem)
}

func (b *Bot) MustClickElementAndWait(elem *rod.Element, opts ...ElemOptionFunc) {
	b.MustClickElement(elem, opts...)
	b.page.MustWaitStable()
}

func (b *Bot) MustClickElement(elem *rod.Element, opts ...ElemOptionFunc) {
	b.e(b.ClickElement(elem, opts...))
}

func (b *Bot) ClickElement(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{handleCoverByEsc: true, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.clickByScript {
		return b.ClickWithScript(elem, opts...)
	}

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	err := b.ensureClickable(elem, opt.handleCoverByEsc)
	if err != nil {
		return err
	}

	err = elem.Timeout(b.shortToSec).Click(proto.InputMouseButtonLeft, 1)
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, &rod.ErrInvisibleShape{}) {
		return b.ClickWithScript(elem, opts...)
	}

	return err
}

func (b *Bot) ensureClickable(elem *rod.Element, handle bool) error {
	err := b.IsInteractable(elem)
	if err == nil {
		return nil
	}

	if hit := b.CloseIfHasPopovers(); hit != 0 {
		log.Trace().Int("hit", hit).Msg("closed pop")
		return nil
	}

	if errors.Is(err, &rod.ErrCovered{}) && handle {
		return b.PressEscape(elem)
	}

	return err
}

func (b *Bot) IsInteractable(elem *rod.Element) error {
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
	err := b.PressEscape(elem)
	b.e(err)
}

func (b *Bot) PressEscape(elem *rod.Element) error {
	return elem.Timeout(b.shortToSec).MustKeyActions().Press(input.Escape).Do()
}

func (b *Bot) ClickWithScript(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{timeout: ShortToSec, highlight: true}
	bindElemOptions(&opt, opts...)

	if opt.highlight {
		b.FocusAndHighlight(elem)
	}

	_, err := elem.Timeout(time.Duration(opt.timeout)*time.Second).Eval(`(elem) => { this.click() }`, elem)
	if err != nil {
		log.Error().Err(err).Msg("Err: close by Eval script this.click()")
		return err
	}

	_, err = elem.Interactable()
	if errors.Is(err, &rod.ErrObjectNotFound{}) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, &rod.ErrInvisibleShape{}) {
		return nil
	}

	return err
}

func (b *Bot) MustAcceptCookies(cookies ...string) {
	for _, sel := range cookies {
		b.MustClickElem(sel)
	}
}

func (b *Bot) CloseIfHasPopovers() int {
	hit := 0

	if len(b.popovers) == 0 {
		return 0
	}

	for _, pop := range b.popovers {
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

	elems, err := b.GetElems(sel)
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

		b.Highlight(elem)

		e := elem.Click(proto.InputMouseButtonLeft, 1)
		if e != nil {
			return 0, e
		}

		hit += 1
	}

	return hit, nil
}
