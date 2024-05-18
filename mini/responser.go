package mini

import "wee"

type Response struct {
	*Shadow

	Bot      *wee.Bot
	pageType pageType
}

func NewResponse(s *Shadow, bot *wee.Bot) *Response {
	return &Response{
		Shadow: s,
		Bot:    bot,
	}
}

func (c *Response) PageData() *string {
	c.Bot.MustWaitLoad()

	var pageData string

	if rps := c.Kwargs.ResponseParseScript; rps != "" {
		pageData = c.Bot.MustEval(c.Kwargs.ResponseParseScript)
		c.pageType = _pageTypeJSON
	} else {
		pageData = c.Bot.Page().MustHTML()
		c.pageType = _pageTypeHTML
	}

	return &pageData
}

func (c *Response) PageType() pageType {
	return c.pageType
}
