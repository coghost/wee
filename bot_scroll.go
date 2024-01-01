package wee

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func (b *Bot) ScrollToElement(elem *rod.Element, opts ...ElemOptionFunc) {
	b.scrollToElement(elem, opts...)
}

func (b *Bot) scrollToElement(elem *rod.Element, opts ...ElemOptionFunc) {
	opt := ElemOptions{offsetToTop: 0.25, steps: ShortScrollStep}
	bindElemOptions(&opt, opts...)

	shape, err := elem.Shape()
	if err != nil {
		return
	}

	box := shape.Box()
	if box == nil {
		return
	}

	h := b.GetWindowInnerHeight()
	scrollDistance := box.Y - h*opt.offsetToTop

	b.page.Mouse.Scroll(0.0, scrollDistance, opt.steps)
}

func (b *Bot) ScrollToElemDirectly(elem *rod.Element) error {
	box, err := b.GetElemBox(elem)
	if err != nil {
		return err
	}

	return b.page.Mouse.Scroll(0.0, box.Y, 1)
}

func (b *Bot) MustScrollToXY(x, y float64) {
	b.page.Mouse.MustScroll(x, y)
}

func (b *Bot) MustScrollToTop() {
	h := b.GetScrollHeight()
	// e := b.page.Mouse.Scroll(0, -h, MediumScrollStep)
	e := b.ScrollLikeHuman(0, -h)
	b.pie(e)
}

func (b *Bot) MustScrollToBottom(opts ...ElemOptionFunc) {
	h := b.GetScrollHeight()
	e := b.ScrollLikeHuman(0, h, opts...)
	b.pie(e)
}

func (b *Bot) ScrollLikeHuman(offsetX, offsetY float64, opts ...ElemOptionFunc) error {
	page := b.page

	scroller := &ScrollAsHuman{
		enabled:          false,
		longSleepChance:  0.1,
		shortSleepChance: 0.2,
		scrollUpChance:   0.9,
	}

	opt := ElemOptions{scrollAsHuman: scroller}
	bindElemOptions(&opt, opts...)

	steps := MediumScrollStep

	if !opt.scrollAsHuman.enabled || steps == 0 {
		err := page.Mouse.Scroll(offsetX, offsetY, steps)

		RandSleep(0.1, 0.2)
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

	for totalScrolled < totalOffsetNeeded {
		yNegative := false
		// handle too slow scroll
		cost := time.Since(startAt).Seconds()
		if cost > tooSlowTimeoutSec {
			err := page.Mouse.Scroll(offsetX, totalOffsetNeeded-totalScrolled, 1)

			RandSleep(0.1, 0.2)
			return err
		}

		chance := rand.Float64()

		if chance < opt.scrollAsHuman.longSleepChance {
			RandSleep(0.5, 0.6)
			continue
		}

		if chance < opt.scrollAsHuman.shortSleepChance {
			RandSleep(0.25, 0.3)
			continue
		}

		distance := rand.Intn(10) + int(base)

		if chance > opt.scrollAsHuman.scrollUpChance {
			yNegative = true
			distance = rand.Intn(20) + int(base)*2
		}

		if v := totalOffsetNeeded - totalScrolled; int(v) < distance {
			distance = int(v)
		}

		if yNegative {
			distance = -distance
		}

		if e := page.Mouse.Scroll(offsetX, float64(distance), steps); e != nil {
			return e
		}

		totalScrolled += float64(distance)
	}

	return nil
}

func (b *Bot) GetWindowInnerHeight() float64 {
	h := b.page.Timeout(b.shortToSec).MustEval(`() => window.innerHeight`).Int()
	// h := b.page.MustGetWindow().Height
	return float64(h)
}

func (b *Bot) GetScrollHeight() float64 {
	h := b.page.Timeout(b.shortToSec).MustEval(`() => document.body.scrollHeight`).Int()
	return float64(h)
}

func (b *Bot) GetElemBox(elem *rod.Element) (*proto.DOMRect, error) {
	shape, err := elem.Shape()
	if err != nil {
		return nil, err
	}

	return shape.Box(), nil
}
