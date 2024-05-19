package mini

import (
	"context"
	"errors"
	"time"

	"wee"

	"github.com/icza/gox/fmtx"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
)

type Request struct {
	*Shadow

	Bot     *wee.Bot
	inputer Inputer
}

func NewRequest(shadow *Shadow, bot *wee.Bot, inputer Inputer) *Request {
	return &Request{Shadow: shadow, Bot: bot, inputer: inputer}
}

func (c *Request) MockOpen(urls ...string) error {
	uri := wee.FirstOrDefault(c.Mapper.Home, urls...)
	return c.Bot.Open(uri)
}

func (c *Request) MockInput(arr ...string) SerpStatus {
	c.Bot.MustAcceptCookies(c.Mapper.Cookies...)

	if err := c.inputer.Input(c.Mapper.Inputs, arr...); err != nil {
		log.Warn().Err(err).Msg("cannot input")
		return SerpCannotInput
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

	if t := c.Kwargs.MaxScrollTime; t != 0 {
		c.scrollToBottom(t, c.Scheme.Ctrl.Humanized)
	} else {
		c.Bot.MustScrollToBottom(wee.WithHumanized(c.Scheme.Ctrl.Humanized))
	}

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

func (c *Request) scrollToBottom(maxScrollTime int, humanized bool) {
	success := make(chan error)

	go func() {
		c.Bot.MustScrollToBottom(wee.WithHumanized(humanized))
		success <- nil
	}()

	select {
	case <-success:
	case <-time.After(time.Duration(maxScrollTime) * time.Second):
		log.Debug().Msg("timeout reached.")
		c.Bot.MustScrollToBottom()
	}
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

func (c *TextInput) Input(selectors []string, arr ...string) error {
	for i, v := range arr {
		if v == "" {
			continue
		}

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

func (c *ClickInput) Input(selectors []string, arr ...string) error {
	for i, sel := range selectors {
		s := fmtx.CondSprintf(sel, arr[i])
		if err := c.bot.Click(s); err != nil {
			return err
		}
	}

	return nil
}
