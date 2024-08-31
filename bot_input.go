package wee

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

var ErrGetTextAfterInput = errors.New("cannot get text after input")

func errGetTextAfterInputError(msg string) error {
	return fmt.Errorf("%w: %s", ErrGetTextAfterInput, msg)
}

func (b *Bot) MustInput(sel, text string, opts ...ElemOptionFunc) string {
	txt, err := b.Input(sel, text, opts...)
	b.pie(err)

	return txt
}

// Input first clear all content, and then input text content.
func (b *Bot) Input(sel, text string, opts ...ElemOptionFunc) (string, error) {
	opt := ElemOptions{submit: false, timeout: PT10Sec, clearBeforeInput: true, endWithEscape: false, humanized: b.humanized}
	bindElemOptions(&opt, opts...)

	if opt.humanized {
		opts = append(opts, WithTimeout(PT20Sec))
	}

	// click the input elem to trigger before input
	_ = b.Click(sel)

	elem, err := b.Elem(sel, opts...)
	if err != nil {
		return "", fmt.Errorf("cannot get elem: %w", err)
	}

	if elem == nil {
		return "", ErrCannotFindSelector(sel + SEP + text)
	}

	if opt.clearBeforeInput {
		// just a best-effort operation.
		err := elem.SelectAllText()
		if err != nil {
			_ = elem.Input("")
		}
	}

	if opt.trigger {
		_ = b.ClickElem(elem)
	}

	// elem = elem.Timeout(time.Second * b.ShortTo).MustSelectAllText().MustInput(text)
	if err := b.typeAsHuman(elem, text, opt.humanized); err != nil {
		return "", fmt.Errorf("cannot input text: %w", err)
	}

	if opt.endWithEscape {
		action, err := elem.KeyActions()
		if err != nil {
			return "", fmt.Errorf("cannot perform escape key action: %w", err)
		}

		_ = action.Press(input.Escape).Do()
	}

	txt, err := elem.Text()
	if err != nil {
		return "", errGetTextAfterInputError(fmt.Sprintf("%v", err))
	}

	if opt.submit {
		SleepPT100Ms()

		action, err := elem.KeyActions()
		if err != nil {
			return "", fmt.Errorf("cannot perform enter key action: %w", err)
		}

		if err := action.Press(input.Enter).Do(); err != nil {
			return "", fmt.Errorf("cannot submit after input text: %w", err)
		}
	}

	return txt, nil
}

// typeAsHuman
//
//	each time before enter (n=args[0] or 5) chars, we wait (to=args[1]/10 or 0.1) seconds
//
//	@return *rod.Element
func (b *Bot) typeAsHuman(elem *rod.Element, text string, humanized bool) error {
	if !humanized {
		err := elem.Input(text)
		if err != nil {
			return fmt.Errorf("cannot input: %w", err)
		}

		return nil
	}

	length := 5
	arr := NewStringSlice(text, length, true)

	for _, str := range arr {
		if err := elem.Timeout(PT20Sec * time.Second).Input(str); err != nil {
			return fmt.Errorf("cannot input by humanized: %w", err)
		}
	}

	wait := 0.1
	RandSleep(wait-0.01, wait+0.01) //nolint:mnd

	return nil
}

func (b *Bot) TypeCharsOneByOne(elem *rod.Element, value string) {
	elem.MustKeyActions().Type([]input.Key(fmt.Sprintf("%v", value))...).MustDo()
}

func (b *Bot) InputSelect(sel, text string, opts ...ElemOptionFunc) error {
	opt := ElemOptions{selectorType: rod.SelectorTypeText}
	bindElemOptions(&opt, opts...)

	elem, err := b.NotNilElem(sel, opts)
	if err != nil {
		return err
	}

	if err := elem.Select([]string{text}, true, opt.selectorType); err != nil {
		return err
	}

	return elem.Blur()
}
