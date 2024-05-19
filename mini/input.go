package mini

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func (c *Crawler) MockInput(arr ...string) SerpStatus {
	st, err := c.getSearchStatus()
	if err == nil && st.PageNum > 0 {
		c.inputByReloadPage(st.URL, st.PageNum)
		return SerpOk
	}

	if err != nil && !errors.Is(err, errNoSearchManger) {
		log.Error().Err(err).Msg("cannot get search status")
	}

	return c.mockInput(arr)
}

func (c *Crawler) inputByReloadPage(url string, pageNum int) error {
	err := c.MockOpen(url)
	if err != nil {
		return err
	}

	c.setPageNum(pageNum)
	log.Debug().Int("page_num", pageNum).Msgf("continue from last")

	return nil
}

func (c *Crawler) mockInput(arr []string) SerpStatus {
	c.bot.MustAcceptCookies(c.scheme.Mapper.Cookies...)

	if err := c.Inputer.Input(c.scheme.Mapper.Inputs, arr...); err != nil {
		log.Warn().Err(err).Msg("cannot input")
		return SerpCannotInput
	}

	if c.scheme.Mapper.Submit != "" {
		c.bot.MustClick(c.scheme.Mapper.Submit)
	}

	return ensureSerpByResults(c.bot.MustAnyElem, c.scheme.Mapper.HasResults, c.scheme.Mapper.NoResults)
}

func (c *Crawler) bindSearch(keywords []string) {
	c.searchTerm = strings.Join(keywords, "_")
	c.searchID = fmt.Sprintf("%d:%s", c.site, c.searchTerm)
}
