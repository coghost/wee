package mini

import (
	"github.com/icza/gox/fmtx"
	"github.com/thoas/go-funk"
)

// ClickSearcher do search all by click
type ClickSearcher struct {
	*Shadow
}

func (c *ClickSearcher) MockOpen() bool {
	c.Bot.MustOpen(c.Mapper.Home)
	c.Bot.MustAcceptCookies(c.Mapper.Cookies...)

	return true
}

func (c *ClickSearcher) MockInput(arr ...string) SerpStatus {
	for i, ft := range c.Mapper.Filters {
		s := fmtx.CondSprintf(ft, arr[i])
		c.Bot.MustClick(s)
	}

	if c.Mapper.Submit != "" {
		c.Bot.MustClick(c.Mapper.Submit)
	}

	return c.ensureSerp()
}

func (c *ClickSearcher) ensureSerp() SerpStatus {
	return c.ensureSerpByResults()
}

func (c *ClickSearcher) ensureSerpByResults() SerpStatus {
	selectors := c.Mapper.HasResults
	selectors = append(selectors, c.Mapper.NoResults...)

	if len(selectors) == 0 {
		return SerpOk
	}

	got := c.Bot.MustAnyElem(selectors...)
	if got == "" {
		return SerpFailed
	}

	if funk.ContainsString(c.Mapper.HasResults, got) {
		return SerpOk
	}

	return SerpNoJobs
}
