package wee

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	_offsetToTop      = 0.25
	_scrollHeight     = 1024
	_longSleepChance  = 0.05
	_shortSleepChance = 0.15
	_scrollUpChance   = 0.9
)

var ErrElemShapeBox = errors.New("cannot get element box by shape.Box()")

// ScrollToElement just a do the best operation
func (b *Bot) ScrollToElement(elem *rod.Element, opts ...ElemOptionFunc) {
	_ = b.scrollToElement(elem, opts...)
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

	return b.Scroll(0.0, scrollDistance, opt.steps)
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
	b.ensureScrollHeight()

	e := b.ScrollLikeHuman(0, -b.scrollHeight)
	b.pie(e)
}

func (b *Bot) ensureScrollHeight() {
	height, err := b.GetScrollHeight()
	if err != nil {
		// if b.scrollHeight == 0 {
		// 	b.pie(err)
		// }
		b.scrollHeight = _scrollHeight
	} else {
		b.scrollHeight = height
	}
}

func (b *Bot) MustScrollToBottom(opts ...ElemOptionFunc) {
	b.ensureScrollHeight()

	e := b.ScrollLikeHuman(0, b.scrollHeight, opts...)
	b.pie(e)
}

func (b *Bot) ScrollToBottom(opts ...ElemOptionFunc) error {
	b.ensureScrollHeight()
	return b.ScrollLikeHuman(0, b.scrollHeight, opts...)
}

func (b *Bot) ScrollLikeHuman(offsetX, offsetY float64, opts ...ElemOptionFunc) error {
	scroller := &ScrollAsHuman{
		enabled:          false,
		longSleepChance:  _longSleepChance,
		shortSleepChance: _shortSleepChance,
		scrollUpChance:   _scrollUpChance,
	}

	opt := ElemOptions{scrollAsHuman: scroller}
	bindElemOptions(&opt, opts...)

	steps := MediumScrollStep
	enabled := opt.scrollAsHuman.enabled || opt.humanized

	if !enabled || steps == 0 {
		err := b.Scroll(offsetX, offsetY, steps)

		SleepPT100Ms()

		return err
	}

	tooSlowTimeoutSec := 20.0
	totalScrolled := 0.0
	totalOffsetNeeded := offsetY

	base := offsetY / float64(steps)

	if offsetY < 0 {
		totalOffsetNeeded = math.Abs(totalOffsetNeeded)
	}

	startAt := time.Now()

	const (
		downPixel = 10
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
			SleepPT500Ms()
			continue
		}

		if chance < opt.scrollAsHuman.shortSleepChance {
			SleepPT250Ms()
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
