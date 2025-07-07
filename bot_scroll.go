package wee

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

const (
	_offsetToTop      = 0.35
	_scrollHeight     = 1024
	_longSleepChance  = 0.05
	_shortSleepChance = 0.15
	_scrollUpChance   = 0.9
)

var ErrElemShapeBox = errors.New("cannot get element box by shape.Box()")

// PressPageDown simulates pressing the Page Down key multiple times on the main document body.
//   - It takes an integer parameter 'times' specifying the number of times to press Page Down.
//   - It takes `html>body` as the element to perform page down
//
// Returns an error if the body element cannot be found or if the key press operation fails.
func (b *Bot) PressPageDown(times int) error {
	elem, err := b.Elem("html>body", WithTimeout(PT60Sec))
	if err != nil {
		return err
	}

	return b.PressPageDownOnElem(elem, times)
}

func (b *Bot) PressPageDownOnElem(elem *rod.Element, times int) error {
	for range times {
		ka, _ := elem.Timeout(b.mediumTimeout).KeyActions()
		if err := ka.Press(input.PageDown).Do(); err != nil {
			return err
		}

		time.Sleep(200 * time.Millisecond) //nolint:mnd
	}

	return nil
}

// ScrollToElement just a do the best operation
func (b *Bot) ScrollToElement(elem *rod.Element, opts ...ElemOptionFunc) {
	_ = b.scrollToElement(elem, opts...)
}

func (b *Bot) ScrollToElementE(elem *rod.Element, opts ...ElemOptionFunc) error {
	return b.scrollToElement(elem, opts...)
}

func (b *Bot) scrollToElement(elem *rod.Element, opts ...ElemOptionFunc) error {
	opt := ElemOptions{offsetToTop: _offsetToTop, steps: ShortScrollStep}
	bindElemOptions(&opt, opts...)

	shape, err := elem.Shape()
	if err != nil {
		return err
	}

	box := shape.Box()
	if box == nil {
		return ErrElemShapeBox
	}

	h := b.GetWindowInnerHeight()
	scrollDistance := box.Y - h*opt.offsetToTop

	// return b.Scroll(0.0, scrollDistance, opt.steps)
	return b.ScrollLikeHuman(0, scrollDistance, opts...)
}

// Scroll Scroll with mouse.
func (b *Bot) Scroll(x, y float64, step int) error {
	return b.page.Mouse.Scroll(x, y, step)
}

func (b *Bot) ScrollToElemDirectly(elem *rod.Element) error {
	box, err := b.GetElemBox(elem)
	if err != nil {
		return err
	}

	return b.Scroll(0.0, box.Y, 1)
}

func (b *Bot) MustScrollToXY(x, y float64) {
	b.page.Mouse.MustScroll(x, y)
}

func (b *Bot) MustScrollToTop() {
	b.tryGetScrollHeight()

	e := b.ScrollLikeHuman(0, -b.scrollHeight)
	b.pie(e)
}

func (b *Bot) tryGetScrollHeight() {
	height, err := b.GetScrollHeight()
	if err != nil {
		// if b.scrollHeight == 0 {
		// 	b.pie(err)
		// }
		b.scrollHeight = _scrollHeight
	} else {
		b.scrollHeight = height - 100
	}
}

func (b *Bot) MustScrollToBottom(opts ...ElemOptionFunc) {
	e := b.ScrollToBottom(opts...)
	b.pie(e)
}

func (b *Bot) ScrollToBottom(opts ...ElemOptionFunc) error {
	b.tryGetScrollHeight()
	return b.ScrollLikeHuman(0, b.scrollHeight, opts...)
}

// TryScrollToBottom just try scroll to bottom, error will be ignored.
func (b *Bot) TryScrollToBottom(opts ...ElemOptionFunc) {
	_ = b.ScrollToBottom(opts...)
}

func (b *Bot) ScrollLikeHuman(offsetX, offsetY float64, opts ...ElemOptionFunc) error {
	scroller := &ScrollAsHuman{
		enabled:          false,
		longSleepChance:  _longSleepChance,
		shortSleepChance: _shortSleepChance,
		scrollUpChance:   _scrollUpChance,
	}

	opt := ElemOptions{scrollAsHuman: scroller, steps: MediumScrollStep}
	bindElemOptions(&opt, opts...)

	steps := opt.steps
	enabled := opt.scrollAsHuman.enabled || opt.humanized

	if !enabled || steps == 0 {
		err := b.Scroll(offsetX, offsetY, steps)

		SleepPT100Ms()

		return err
	}

	tooSlowTimeoutSec := 60.0
	totalScrolled := 0.0
	totalOffsetNeeded := offsetY

	base := offsetY / float64(steps)

	if offsetY < 0 {
		totalOffsetNeeded = math.Abs(totalOffsetNeeded)
	}

	startAt := time.Now()

	const (
		downPixel = 8
	)

	for totalScrolled < totalOffsetNeeded {
		yNegative := false
		// handle too slow scroll
		cost := time.Since(startAt).Seconds()
		if cost > tooSlowTimeoutSec {
			err := b.Scroll(offsetX, totalOffsetNeeded-totalScrolled, 1)

			SleepPT100Ms()

			return err
		}

		chance := rand.Float64()

		if chance < opt.scrollAsHuman.longSleepChance {
			RandSleepNap()
			continue
		}

		if chance < opt.scrollAsHuman.shortSleepChance {
			SleepPT500Ms()
			continue
		}

		distance := rand.Intn(downPixel) + int(base)

		if chance > opt.scrollAsHuman.scrollUpChance {
			yNegative = true
			distance = rand.Intn(downPixel*2) + int(base)*2 //nolint:mnd
		}

		if v := totalOffsetNeeded - totalScrolled; int(v) < distance {
			distance = int(v)
		}

		if yNegative {
			distance = -distance
		}

		if e := b.Scroll(offsetX, float64(distance), steps); e != nil {
			return e
		}

		totalScrolled += float64(distance)
	}

	return nil
}

func (b *Bot) GetWindowInnerHeight() float64 {
	h := b.page.Timeout(b.shortTimeout).MustEval(`() => window.innerHeight`).Int()
	// h := b.page.MustGetWindow().Height
	return float64(h)
}

func (b *Bot) GetScrollHeight() (float64, error) {
	res, err := b.page.Timeout(b.shortTimeout).Eval(`() => document.body.scrollHeight`)
	if err != nil {
		return 0, err
	}

	return res.Value.Num(), nil
}

func (b *Bot) GetElemBox(elem *rod.Element) (*proto.DOMRect, error) {
	shape, err := elem.Shape()
	if err != nil {
		return nil, err
	}

	return shape.Box(), nil
}
