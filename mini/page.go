package mini

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wee"

	"github.com/rs/zerolog/log"
)

func (c *Crawler) MockOpen(urls ...string) error {
	uri := wee.FirstOrDefault(c.scheme.Mapper.Home, urls...)
	// in case bot is not launched.
	wee.BindBotLanucher(c.bot)

	return c.bot.Open(uri)
}

func (c *Crawler) gotoNextPageInTime(pageNum int) bool {
	success := make(chan error)

	go func() {
		success <- c.gotoNextPage()
	}()

	select {
	case err := <-success:
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				log.Error().Err(err).Int("page_num", pageNum).Msg("get serp failed")
			}

			// log.Error().Err(err).Int("page_num", pageNum).Msg("get serp failed")
			return false
		}
	case <-time.After(time.Duration(_perPageMaxDwellSeconds) * time.Second):
		log.Error().Int("page_num", pageNum).Msg("get serp timeout")
		return false
	}

	return true
}

func (c *Crawler) gotoNextPage() error {
	// 1. wait page is loaded
	elem, err := c.bot.Elem(c.scheme.Mapper.NextPage, wee.WithTimeout(wee.NapToSec))
	if err != nil {
		return fmt.Errorf("cannot get button: %w", err)
	}

	return c.bot.ClickElem(elem)
}

func (c *Crawler) scrollDownToBottom() {
	if c.scheme.Kwargs.MaxPages == c.pageNum {
		c.bot.MustScrollToBottom(wee.WithHumanized(false)) // no humanize in final page.
	}

	if t := c.scheme.Kwargs.MaxScrollTime; t != 0 {
		c.scrollToBottomInTime(t, c.scheme.Ctrl.Humanized)
	} else {
		c.bot.MustScrollToBottom(wee.WithHumanized(c.scheme.Ctrl.Humanized))
	}
}

func (c *Crawler) scrollToBottomInTime(maxScrollTime int, humanized bool) {
	success := make(chan error)

	go func() {
		c.bot.MustScrollToBottom(wee.WithHumanized(humanized))
		success <- nil
	}()

	select {
	case <-success:
	case <-time.After(time.Duration(maxScrollTime) * time.Second):
		log.Debug().Msg("scroll timeout.")
		c.bot.MustScrollToBottom()
	}
}

func (c *Crawler) waitInPagination() {
	itv := c.scheme.Kwargs.PageInterval
	if itv <= 0 {
		return
	}

	wee.RandSleep(itv, itv+1)
}
