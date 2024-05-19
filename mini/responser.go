package mini

import "wee"

type Response struct {
	*Shadow

	Bot      *wee.Bot
	pageType PageType
}

func NewResponse(s *Shadow, bot *wee.Bot) *Response {
	return &Response{
		Shadow: s,
		Bot:    bot,
	}
}

func (c *Response) PageData() (PageType, *string) {
	c.Bot.MustWaitLoad()

	var pageData string

	if rps := c.Kwargs.ResponseParseScript; rps != "" {
		pageData = c.Bot.MustEval(c.Kwargs.ResponseParseScript)
		c.pageType = PageTypeJSON
	} else {
		pageData = c.Bot.Page().MustHTML()
		c.pageType = PageTypeHTML
	}

	return c.pageType, &pageData
}

func (c *Response) PageType() PageType {
	return c.pageType
}
