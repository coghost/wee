package wee

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

func (b *Bot) MustInput(sel, text string, opts ...ElemOptionFunc) string {
	txt, err := b.Input(sel, text, opts...)
	b.e(err)

	return txt
}

// Input first clear all content, and then input text content.
func (b *Bot) Input(sel, text string, opts ...ElemOptionFunc) (string, error) {
	opt := ElemOptions{submit: false}
	bindElemOptions(&opt, opts...)

	elem, err := b.GetElem(sel, opts...)
	if err != nil {
		return "", err
	}

	if elem == nil {
		return "", ErrCannotFindElem(sel + "@@@" + text)
	}

	// elem = elem.Timeout(time.Second * b.ShortTo).MustSelectAllText().MustInput(text)
	b.typeAsHuman(elem, text, opt.humanized)

	if opt.submit {
		RandSleep(0.1, 0.15)
		if err := elem.MustKeyActions().Press(input.Enter).Do(); err != nil {
			return "", err
		}
	}

	// just try to get text, won't matter if fails
	return elem.Text()
}

// typeAsHuman
//
//	each time before enter (n=args[0] or 5) chars, we wait (to=args[1]/10 or 0.1) seconds
//
//	@return *rod.Element
func (b *Bot) typeAsHuman(elem *rod.Element, text string, humanized bool) {
	elem.MustSelectAllText().MustInput("")

	if !humanized {
		elem.MustInput(text)
		return
	}

	length := 5
	wait := 0.1

	arr := NewStringSlice(text, length, true)

	for _, str := range arr {
		err := elem.Input(str)
		b.e(err)
	}

	RandSleep(wait-0.01, wait+0.01)
}

func (b *Bot) TypeCharsOneByOne(elem *rod.Element, value string) {
	elem.MustKeyActions().Type([]input.Key(fmt.Sprintf("%v", value))...).MustDo()
}
