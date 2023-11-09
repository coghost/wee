package wee

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func (b *Bot) FocusAndHighlight(elem *rod.Element) {
	if b.highlightTimes == 0 {
		b.e(elem.Focus())
	} else {
		b.ScrollToElement(elem)
		b.Highlight(elem)
	}
}

func (b *Bot) Highlight(elem *rod.Element) {
	if b.highlightTimes == 0 {
		return
	}

	show, hide := 0.333, 0.2
	style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`

	go b.highlight(elem, show, hide, style, 0)
}

func (b *Bot) highlight(elem *rod.Element, show, hide float64, style string, count int) float64 {
	start := time.Now()

	if b.highlightTimes == 0 {
		return 0
	}

	if elem == nil {
		return 0
	}

	ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	if err != nil {
		log.Debug().Msg("No style found")
		return 0
	}

	origStyle := ob.Value.String()
	// origStyle := elem.MustEval(`e => {return this.getAttribute("style")}`).String()
	// style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	// show, hide := 0.333, 0.2
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
