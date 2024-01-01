package mini

import (
	"context"
	"errors"

	"wee"

	"github.com/rs/zerolog/log"
)

type Response struct {
	*Shadow
}

func (c *Response) GotoNextPage() error {
	c.Bot.MustWaitLoad()
	c.Bot.MustScrollToBottom()

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

func (c *Response) ResponsePageData() *string {
	c.Bot.MustWaitLoad()

	var pageData string

	if rps := c.Kwargs.ResponseParseScript; rps != "" {
		pageData = c.Bot.MustEval(c.Kwargs.ResponseParseScript)
	} else {
		pageData = c.Bot.Page().MustHTML()
	}

	return &pageData
}

func (c *Response) MockHumanWait() {
	if itv := c.Kwargs.PageInterval; itv > 0 {
		wee.RandSleep(itv, itv+0.5)
	}
}
