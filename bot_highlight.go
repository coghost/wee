package wee

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

const (
	_zoomIn = `zoom: 1.25;-moz-transform: scale(1.25);`
	_style  = `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	_style1 = `box-shadow: 0 0 10px rgba(255,125,0,1), 0 0 20px 5px rgba(255,175,0,0.8), 0 0 30px 15px rgba(255,225,0,0.5);`
)

func (b *Bot) FocusAndHighlight(elem *rod.Element) {
	if b.highlightTimes == 0 {
		elem.Focus()
	} else {
		b.ScrollToElement(elem)
		b.HighlightElem(elem)
	}
}

func (b *Bot) HighlightElem(elem *rod.Element) {
	if b.highlightTimes == 0 {
		return
	}

	show, hide := 0.333, 0.2

	go b.Highlight(elem, show, hide, _style1, 0)
}

func (b *Bot) Highlight(elem *rod.Element, show, hide float64, style string, count int) float64 {
	start := time.Now()

	if b.highlightTimes == 0 {
		return 0
	}

	if elem == nil {
		return 0
	}

	origStyle := ""

	ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	if err == nil {
		origStyle = ob.Value.String()
	}

	// origStyle := elem.MustEval(`e => {return this.getAttribute("style")}`).String()
	// style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	base := 0.05

	if count == 0 {
		count = b.highlightTimes
	}

	for i := 0; i < count; i++ {
		script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, style)
		elem.Eval(script)
		RandSleep(show-base, show+base)

		script = fmt.Sprintf(`() => this.setAttribute("style", "%s");`, origStyle)
		elem.Eval(script)
		RandSleep(hide-base, hide+base)
	}

	cost := time.Since(start).Seconds()

	return cost
}

func (b *Bot) MarkElems(timeout time.Duration, elems ...*rod.Element) {
	style := `box-shadow: 0px 0px 0px 3px rgba(148,0,211,1);`
	for _, e := range elems {
		go b.MarkElem(e, timeout, style)
	}
}

func (b *Bot) MarkElem(elem *rod.Element, timeout time.Duration, style string) {
	origStyle := ""

	// ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	// if err == nil {
	// 	origStyle = ob.Value.String()
	// }

	defer func() {
		script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, origStyle)
		elem.Eval(script)
	}()

	if style == "" {
		style = _style
	}

	script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, style)
	elem.Eval(script)

	time.Sleep(timeout)
}
