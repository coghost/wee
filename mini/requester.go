package mini

import (
	"context"
	"errors"

	"wee"

	"github.com/icza/gox/fmtx"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
)

type Request struct {
	*Shadow
	inputer Inputer
}

type Inputer interface {
	Input(bot *wee.Bot, selectors []string, arr ...string) error
}

func NewRequest(shadow *Shadow, inputer Inputer) *Request {
	return &Request{Shadow: shadow, inputer: inputer}
}

func (c *Request) MockOpen() bool {
	c.Bot.MustOpen(c.Mapper.Home)
	c.Bot.MustAcceptCookies(c.Mapper.Cookies...)

	return true
}

func (c *Request) MockInput(arr ...string) SerpStatus {
	if !c.MockOpen() {
		return SerpFailed
	}

	if err := c.inputer.Input(c.Bot, c.Mapper.Inputs, arr...); err != nil {
		return SerpFailed
	}

	if c.Mapper.Submit != "" {
		c.Bot.MustClick(c.Mapper.Submit)
	}

	return c.ensureSerp()
}

func (c *Request) ensureSerp() SerpStatus {
	return ensureSerpByResults(c.Bot.MustAnyElem, c.Mapper.HasResults, c.Mapper.NoResults)
}

func (c *Request) GotoNextPage() error {
	c.Bot.MustWaitLoad()
	c.Bot.MustScrollToBottom(wee.WithHumanized(true))

	elem, err := c.Bot.Elem(c.Mapper.NextPage, wee.WithTimeout(wee.NapToSec))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Info().Msg("final page reached")
		} else {
			log.Error().Err(err).Msg("cannot find next page button")
		}

		return err
	}

	return c.Bot.ClickElem(elem)
}

func (c *Request) MockHumanWait() {
	if itv := c.Kwargs.PageInterval; itv > 0 {
		wee.RandSleep(itv, itv+0.5)
	}
}

func ensureSerpByResults(ensure func(selectors ...string) string, hasResults []string, noResults []string) SerpStatus {
	selectors := hasResults
	selectors = append(selectors, noResults...)

	if len(selectors) == 0 {
		return SerpOk
	}

	got := ensure(selectors...)
	if got == "" {
		return SerpFailed
	}

	if funk.ContainsString(hasResults, got) {
		return SerpOk
	}

	return SerpNoJobs
}

type TextInput struct {
	bot *wee.Bot
}

func newTextInput(bot *wee.Bot) *TextInput {
	return &TextInput{bot: bot}
}

func (c *TextInput) Input(bot *wee.Bot, selectors []string, arr ...string) error {
	for i, v := range arr {
		_, err := c.bot.Input(selectors[i], v)
		if err != nil {
			return err
		}
	}

	return nil
}

type ClickInput struct {
	bot *wee.Bot
}

func (c *ClickInput) Input(bot *wee.Bot, selectors []string, arr ...string) error {
	for i, sel := range selectors {
		s := fmtx.CondSprintf(sel, arr[i])
		if err := c.bot.Click(s); err != nil {
			return err
		}
	}

	return nil
}
