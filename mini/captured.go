package mini

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"wee"

	"github.com/rs/zerolog/log"
)

func (c *Crawler) parseResultsCount() (string, float64) {
	sel := c.scheme.Mapper.ResultsCount
	if sel == "" {
		return "", 0
	}

	rc := c.scheme.Kwargs.ResultsCount
	txt := c.bot.MustElemAttr(sel, wee.WithAttr(rc.Attr), wee.WithIndex(rc.Index))

	if rc.AttrSep != "" {
		txt = strings.Split(txt, rc.AttrSep)[rc.AttrIndex]
	}

	return txt, wee.MustStrToFloat(txt, rc.CharsAllowed)
}

var errNoData = errors.New("no data got")

func (c *Crawler) saveCaptured() (*SerpDirective, error) {
	pageType, content := c.getPageData()

	if content == nil {
		return nil, errNoData
	}

	filePath := c.genObjPath(pageType)

	if err := c.FileStorage.PutObject(filePath, []byte(*content)); err != nil {
		// log.Error().Err(err).Str("filepath", filePath).Msg("cannot save captured data")
		return nil, fmt.Errorf("cannot save data: %w", err)
	}

	return c.newSerpDirective(filePath, content), nil
}

func (c *Crawler) genObjPath(pageType PageType) string {
	name := fmt.Sprintf("%s_%s_%d", c.SerpIdNowFunc(), c.searchTerm, c.pageNum)
	name = strings.ReplaceAll(name, "/", "-")
	filePath := fmt.Sprintf("%d/%s.%s", c.site, name, pageType)

	return filePath
}

func (c *Crawler) getPageData() (PageType, *string) {
	c.bot.MustWaitLoad()

	var (
		pageType PageType
		pageData string
	)

	if script := c.scheme.Kwargs.ResponseParseScript; script != "" {
		pageData = c.bot.MustEval(script)
		pageType = PageTypeJSON
	} else {
		pageData = c.bot.Page().MustHTML()
		pageType = PageTypeHTML
	}

	return pageType, &pageData
}

func (c *Crawler) newSerpDirective(filepath string, raw *string) *SerpDirective {
	serpId := fmt.Sprintf("%s_%s", c.SerpIdNowFunc(), c.searchID)

	return &SerpDirective{
		Site:     c.site,
		PageNum:  c.pageNum,
		PageSize: len(*raw),

		ResultsLoaded: c.getResultsLoaded(),
		ResultsCount:  c.serpResultsCount,

		PageDwellTime: time.Since(c.startTime).Seconds(),
		Filepath:      filepath,
		SerpUrl:       c.bot.CurrentUrl(),
		SearchTerm:    c.searchTerm,
		SerpID:        serpId,
		SearchID:      c.searchID,
	}
}

func (c *Crawler) getResultsLoaded() int {
	count := 0

	for _, s := range c.scheme.Mapper.HasResults {
		elems, err := c.bot.Elems(s)
		if err != nil {
			log.Warn().Str("selector", s).Msg("cannot get elems")
			continue
		}

		count += len(elems)
	}

	c.pageResultsLoaded = count

	return count
}
